package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sfomuseum/go-sfomuseum-curatorial/render"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"net/url"
	"sync"
)

func main() {

	iterator_uri := flag.String("iterator-uri", "", "...")
	iterator_source := flag.String("iterator-source", "", "...")

	reader_uri := flag.String("reader-uri", "", "")
	parent_reader_uri := flag.String("parent-reader-uri", "", "")

	outdir := flag.String("outdir", "/usr/local/data/sfomuseum-data-architecture/sfomuseum/galleries", "...")

	flag.Parse()

	ctx := context.Background()

	abs_root, err := filepath.Abs(*outdir)

	if err != nil {
		log.Fatalf("Failed to derive absolute path for %s, %v", *outdir, err)
	}

	feature_r, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new reader, %v", err)
	}

	parent_r, err := reader.NewReader(ctx, *parent_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new parent reader, %v", err)
	}

	galleries := make([]int64, 0)
	lookup := new(sync.Map)

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		rsp := gjson.GetBytes(body, "properties.sfomuseum:gallery_id")

		if !rsp.Exists() {
			return nil
		}

		lookup.Store(rsp.Int(), true)
		return nil
	}

	iter, err := iterator.NewIterator(ctx, *iterator_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to iterate source, %v", err)
	}

	lookup.Range(func(k interface{}, v interface{}) bool {
		galleries = append(galleries, k.(int64))
		return true

	})

	for _, gallery_id := range galleries {

		// START OF need to fix json-query stuff
		// to decode URL parameters

		u, _ := url.Parse(*iterator_uri)

		q := u.Query()
		q.Set("include", fmt.Sprintf("include=properties.sfomuseum:gallery_id=^%d$", gallery_id))

		u.RawQuery = q.Encode()

		gallery_iterator_uri := u.String()

		// END OF need to fix json-query stuff

		gallery_iterator_uri = fmt.Sprintf("%s&include=properties.sfomuseum:gallery_id=^%d$", *iterator_uri, gallery_id)

		fname := fmt.Sprintf("gallery-%03d.png", gallery_id)
		path := filepath.Join(abs_root, fname)

		opts := &render.RenderOptions{
			FeatureReader:   feature_r,
			ParentReader:    parent_r,
			IteratorURI:     gallery_iterator_uri,
			IteratorSources: []string{*iterator_source},
		}

		log.Println(path)

		_, body, err := render.Render(ctx, opts)

		if err != nil {
			log.Fatal(err)
		}

		wr, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Failed to create writer for %s, %v", path, err)
		}

		err = render.Draw(ctx, body, wr)

		if err != nil {
			log.Fatalf("Failed to draw %s, %v", path, err)
		}

		err = wr.Close()

		if err != nil {
			log.Fatalf("Failed to close writer for %s, %v", path, err)
		}
	}

}
