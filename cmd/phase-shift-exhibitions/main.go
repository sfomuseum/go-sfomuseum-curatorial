// Update and supersese all the exhibition records to reflect new gallery parentage after a complex-level "phase shift".
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v3"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"	
	"github.com/whosonfirst/go-whosonfirst-id"
	"github.com/whosonfirst/go-writer/v3"
)

func main() {

	var architecture_reader_uri string
	var exhibitions_reader_uri string
	var exhibitions_writer_uri string	

	var exhibitions_iterator_uri string
	var exhibitions_iterator_source string	
	
	flag.StringVar(&architecture_reader_uri, "architecture-reader-uri", "repo:///usr/local/data/sfomuseum-data-architecture", "")
	flag.StringVar(&exhibitions_reader_uri"exhibitions-reader-uri", "repo:///usr/local/data/sfomuseum-data-exhibition", "")
	
	flag.StringVar(&exhibitions_writer_uri, "exhibitions-writer-uri", "", "If empty, the value of the -exhibition-reader-uri flag will be used.")

	flag.StringVar(&exhibitions_iterator_uri, "exhibitions-iterator-uri", "repo://?include=properties.mz:is_current=1", "")
	flag.StringVar(&exhibitions_iterator_source, "exhibitions-iterator-source", "/usr/local/data/sfomuseum-data-exhibition", "")
	
	flag.Parse()

	ctx := context.Background()

	if exhibition_writer_uri == "" {
		*exhibition_writer_uri = exhibition_reader_uri
	}

	arch_r, err := reader.NewReader(ctx, architecture_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create architecture reader, %v", err)
	}

	// START OF build spatial/PIP stuff (put me in a package...)


	// END OF build spatial/PIP stuff (put me in a package...)	
	
	exh_r, err := reader.NewReader(ctx, exhibition_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create exhibition reader, %v", err)
	}

	exh_wr, err := writer.NewWriter(ctx, exhibition_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create exhibition writer, %v", err)
	}

	exhibitions_iterator_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		// PIP body here, derive new parent_id

		return nil
	}

	exhibitions_iterator, err := iterator.NewIterator(ctx, exhibitions_iterator_uri, exhibitions_iterator_cb)

	if err != nil {
		log.Fatalf("Failed to create exhibitions iterator, %v", err)
	}

	err = exhibitions_iterator.IterateURIs(ctx, exhibitions_iterator_source)

	if err != nil {
		log.Fatalf("Failed to iterate %s, %w", exhibitions_iterator_source, err)
	}
}

/*

	
	supersede := func(ctx context.Context, exhibition_id int64, parent_id int64) error {

		id_provider, err := id.NewProvider(ctx)

		if err != nil {
			return fmt.Errorf("Failed to create ID provider, %w", err)
		}

		exh_f, err := sfom_reader.LoadBytesFromID(ctx, exh_r, exhibition_id)

		if err != nil {
			return fmt.Errorf("Failed to load exhibition record, %w", err)
		}

		new_parent_f, err := sfom_reader.LoadBytesFromID(ctx, arch_r, parent_id)

		if err != nil {
			return fmt.Errorf("Failed to load parent record, %w", err)
		}

		new_id, err := id_provider.NewID(ctx)

		if err != nil {
			return fmt.Errorf("Failed to create new ID, %w", err)
		}

		new_updates := map[string]interface{}{
			"properties.id":             new_id,
			"properties.wof:id":         new_id,
			"properties.wof:parent_id":  parent_id,
			"properties.wof:hierarchy":  gjson.GetBytes(new_parent_f, "properties.wof:hierarchy").Value(),
			"properties.mz:is_current":  gjson.GetBytes(new_parent_f, "properties.mz:is_current").Value(),
			"properties.edtf:inception": gjson.GetBytes(new_parent_f, "properties.edtf:inception").Value(),
			// "properties.edtf:cessation": gjson.GetBytes(new_parent_f, "properties.edtf:cessation").Value(),
			"properties.wof:supersedes": []int64{exhibition_id},
		}

		// Create and record the new exh

		_, new_exh, err := export.AssignPropertiesIfChanged(ctx, exh_f, new_updates)

		if err != nil {
			return fmt.Errorf("Failed to export new exh, %w", err)
		}

		_, err = sfom_writer.WriteBytes(ctx, exh_wr, new_exh)

		if err != nil {
			return fmt.Errorf("Failed to write new exhibition, %w", err)
		}

		//

		old_updates := map[string]interface{}{
			"properties.wof:superseded_by": []int64{new_id},
			"properties.edtf:cessation":    gjson.GetBytes(new_parent_f, "properties.edtf:inception").Value(),
		}

		// Now update the previous exh

		_, exh_f, err = export.AssignPropertiesIfChanged(ctx, exh_f, old_updates)

		if err != nil {
			return fmt.Errorf("Failed to export new exhibition, %w", err)
		}

		_, err = sfom_writer.WriteBytes(ctx, exh_wr, exh_f)

		if err != nil {
			return fmt.Errorf("Failed to write previous exhibition, %w", err)
		}

		log.Printf("Created new exhibition with ID %d\n", new_id)

		// Something something something.
		// Do this recursively based on the end date of the exhibition
		// relative to the end date of the parent (architecture).
		// Something something something

		return nil
	}

	err = supersede(ctx, *exhibition_id, *parent_id)

	if err != nil {
		log.Fatalf("Failed to supersede exhibition (%d), %v", *exhibition_id, err)
	}

*/
