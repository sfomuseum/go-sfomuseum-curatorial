package main

// This tool attempts to derive the most appropriate gallery for an exhibition
// based on an exhibition's (sfo museum) gallery ID and the edtf:inception date
// for that exhibition's WOF feature. The tool works mostly (I think) but there
// are still many instances where it can't resolve mulitple galleries where all
// the candidates overlap a given date but are "fuzzy" EDTF dates so it's not
// possible to filter things.

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"sync"

	"github.com/sfomuseum/go-sfomuseum-architecture/galleries"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v3"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
)

func main() {

	var iterator_uri string
	var iterator_source string

	var architecture_reader_uri string
	var exhibitions_writer_uri string

	var verbose bool

	flag.StringVar(&iterator_uri, "iterator-uri", "repo://?exclude=properties.edtf:deprecated=.*", "...")
	flag.StringVar(&iterator_source, "iterator-source", "/usr/local/data/sfomuseum-data-exhibition", "...")

	flag.StringVar(&architecture_reader_uri, "architecture-reader-uri", "repo:///usr/local/data/sfomuseum-data-architecture", "...")
	flag.StringVar(&exhibitions_writer_uri, "exhibitions-writer-uri", "repo:///usr/local/data/sfomuseum-data-exhibition", "...")
	flag.BoolVar(&verbose, "verbose", false, "...")

	flag.Parse()

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	/*
		loc, err := time.LoadLocation("America/Los_Angeles")

		if err != nil {
			log.Fatal(err)
		}
	*/

	architecture_reader, err := reader.NewReader(ctx, architecture_reader_uri)

	if err != nil {
		log.Fatal(err)
	}

	exhibitions_writer, err := writer.NewWriter(ctx, exhibitions_writer_uri)

	if err != nil {
		log.Fatal(err)
	}

	features := new(sync.Map)

	load_feature := func(ctx context.Context, id int64) ([]byte, error) {

		v, exists := features.Load(id)

		if exists {
			return v.([]byte), nil
		}

		body, err := wof_reader.LoadBytes(ctx, architecture_reader, id)

		if err != nil {
			return nil, fmt.Errorf("Failed to load %d, %w", id, err)
		}

		features.Store(id, body)
		return body, nil
	}

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		is_alt, err := uri.IsAltFile(path)

		if err != nil {
			return err
		}

		if is_alt {
			return nil
		}

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		logger := slog.Default()
		logger = logger.With("path", path)

		old_parent_id, err := properties.ParentId(body)

		if err != nil {
			return fmt.Errorf("Failed to derive old parent ID for %s, %w", err)
		}

		logger = logger.With("old parent id", old_parent_id)

		exhibition_date := properties.Inception(body)
		logger = logger.With("date", exhibition_date)

		galleries_rsp := gjson.GetBytes(body, "properties.sfomuseum:gallery_id")

		gallery_ids := make([]int64, 0)

		for _, r := range galleries_rsp.Array() {

			gallery_id := r.Int()
			gallery_id_str := r.String()

			g, err := galleries.FindGalleryForDate(ctx, gallery_id_str, exhibition_date)

			if err != nil {
				logger.Error("Failed to resolve gallery, skipping", "gallery_id", gallery_id, "error", err)
				return nil
			}

			gallery_ids = append(gallery_ids, g.WhosOnFirstId)
		}

		var parent_id int64

		switch len(gallery_ids) {
		case 0:
			parent_id = -1
		case 1:
			parent_id = gallery_ids[0]
		default:
			parent_id = -4
		}

		if parent_id == -1 {
			return nil
		}

		if parent_id == old_parent_id {
			logger.Debug("No change to parent ID, skipping")
			return nil
		}

		logger.Info("Resolved gate ID", "id", parent_id)

		updates := map[string]interface{}{
			"properties.wof:parent_id": parent_id,
		}

		hierarchies := make([]map[string]int64, 0)
		hierarchies_key := new(sync.Map)

		for _, gallery_id := range gallery_ids {

			gallery_body, err := load_feature(ctx, gallery_id)

			if err != nil {
				slog.Warn("Failed to lookup parent", "gallery", gallery_id, "error", err)
			} else {

				gallery_hierarchies := properties.Hierarchies(gallery_body)

				for _, h := range gallery_hierarchies {

					enc_h, err := json.Marshal(h)

					if err != nil {
						slog.Error("Failed to encode hierarchy, skipping", "error", err)
						continue
					}

					sum_h := sha256.Sum256(enc_h)
					hash_h := fmt.Sprintf("%x", sum_h)

					_, exists := hierarchies_key.LoadOrStore(hash_h, true)

					if exists {
						continue
					}

					hierarchies = append(hierarchies, h)
				}
			}
		}

		updates["properties.wof:hierarchy"] = hierarchies

		// Update geom?

		has_changed, new_body, err := export.AssignPropertiesIfChanged(ctx, body, updates)

		if err != nil {
			slog.Error("Failed to assign updates", "error", err)
			return nil
		}

		if !has_changed {
			return nil
		}

		_, err = sfom_writer.WriteBytes(ctx, exhibitions_writer, new_body)

		if err != nil {
			slog.Error("Failed to write updates", "error", err)
			return nil
		}

		logger.Info("Update record", "new parent_id", parent_id)
		return nil
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		log.Fatal(err)
	}

	err = iter.IterateURIs(ctx, iterator_source)

	if err != nil {
		log.Fatal(err)
	}
}
