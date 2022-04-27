// assign-exhibition-gallery is a command line tool to update wof:parent_id and wof:hierarchy information
// for a SFO Museum exhibition record derived from one or more SFO Museum gallery records.
package main

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-flags/multi"
	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-writer"
	"log"
)

func main() {

	architecture_reader_uri := flag.String("architecture-reader-uri", "repo:///usr/local/data/sfomuseum-data-architecture", "")
	exhibitions_reader_uri := flag.String("exhibitions-reader-uri", "repo:///usr/local/data/sfomuseum-data-exhibition", "")
	exhibitions_writer_uri := flag.String("exhibitions-writer-uri", "", "If empty, the value of the -exhibitions-reader-uri flag will be used.")
	exhibition_id := flag.Int64("exhibition-id", 0, "The SFO Museum exhibition ID to update.")

	var gallery_ids multi.MultiInt64
	flag.Var(&gallery_ids, "gallery-id", "One or more SFO Museum gallery IDs.")

	flag.Parse()

	ctx := context.Background()

	if *exhibitions_writer_uri == "" {
		*exhibitions_writer_uri = *exhibitions_reader_uri
	}

	arch_r, err := reader.NewReader(ctx, *architecture_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create architecture reader, %v", err)
	}

	exh_r, err := reader.NewReader(ctx, *exhibitions_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create exhibitions reader, %v", err)
	}

	exh_wr, err := writer.NewWriter(ctx, *exhibitions_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create exhibitions writer, %v", err)
	}

	exh_f, err := sfom_reader.LoadBytesFromID(ctx, exh_r, *exhibition_id)

	if err != nil {
		log.Fatalf("Failed to load exhibition record, %v", err)
	}

	galleries := make([][]byte, len(gallery_ids))

	for idx, gal_id := range gallery_ids {

		gal_f, err := sfom_reader.LoadBytesFromID(ctx, arch_r, gal_id)

		if err != nil {
			log.Fatalf("Failed to load gallery record %d, %v", gal_id, err)
		}

		galleries[idx] = gal_f
	}

	updates := make(map[string]interface{})

	switch len(galleries) {
	case 0:
		updates["properties.wof:parent_id"] = -1
		updates["properties.wof:hierarchy"] = make([]map[string]int64, 0)
	case 1:
		updates["properties.wof:parent_id"] = gjson.GetBytes(galleries[0], "properties.wof:id").Int()
		updates["properties.wof:hierarchy"] = gjson.GetBytes(galleries[0], "properties.wof:hierarchy").Value()
		updates["properties.sfomuseum:post_security"] = gjson.GetBytes(galleries[0], "properties.sfomuseum:post_security").Value()
		updates["geometry"] = gjson.GetBytes(galleries[0], "geometry").Value()
	default:

		hiers := make([]map[string]interface{}, 0)
		coords := make([][]float64, 0)

		for _, body := range galleries {

			for _, r := range gjson.GetBytes(body, "properties.wof:hierarchy").Array() {
				hiers = append(hiers, r.Value().(map[string]interface{}))
			}

			pt, _, err := properties.Centroid(body)

			if err != nil {
				log.Fatalf("Failed to derive centroid for gallery, %v", err)
			}

			coords = append(coords, []float64{pt.X(), pt.Y()})
		}

		updates["properties.wof:parent_id"] = -4
		updates["properties.wof:hierarchy"] = hiers

		updates["geometry.type"] = "MultiPoint"
		updates["geometry.coordinates"] = coords
	}

	has_updates, exh_f, err := export.AssignPropertiesIfChanged(ctx, exh_f, updates)

	if err != nil {
		log.Fatalf("Failed to assign properties, %v", err)
	}

	if has_updates {

		_, err := sfom_writer.WriteFeatureBytes(ctx, exh_wr, exh_f)

		if err != nil {
			log.Fatalf("Failed to write updates, %v", err)
		}
	}
}
