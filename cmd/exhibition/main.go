package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func main() {

	iterator_uri := flag.String("iterator-uri", "", "...")
	iterator_source := flag.String("iterator-source", "", "...")

	parent_reader_uri := flag.String("parent-reader-uri", "", "")

	flag.Parse()

	ctx := context.Background()

	parent_r, err := reader.NewReader(ctx, *parent_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new parent reader, %v", err)
	}

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse '%s', %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		name, err := properties.Name(body)

		if err != nil {
			return fmt.Errorf("Failed to derive name for %s, %w", path, err)
		}

		inception := properties.Inception(body)
		cessation := properties.Cessation(body)

		label := fmt.Sprintf("%d %s %s - %s", id, name, inception, cessation)
		fmt.Println(label)

		parent_id, err := properties.ParentId(body)

		if err != nil {
			return fmt.Errorf("Failed to derive parent ID for %s, %w", path, err)
		}

		parent_body, err := wof_reader.LoadBytes(ctx, parent_r, parent_id)

		if err != nil {
			return fmt.Errorf("Failed to load parent ID (%d) for %s, %w", parent_id, path, err)
		}

		parent_name, err := properties.Name(parent_body)

		if err != nil {
			return fmt.Errorf("Failed to derive parent name (%s) for %s, %w", parent_id, path, err)
		}

		parent_inception := properties.Inception(parent_body)
		parent_cessation := properties.Cessation(parent_body)

		parent_label := fmt.Sprintf("%d %s %s - %s", parent_id, parent_name, parent_inception, parent_cessation)
		fmt.Println(parent_label)

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
}
