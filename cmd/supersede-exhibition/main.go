package main

import (
	"context"
	"flag"
	"fmt"
	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-id"
	"github.com/whosonfirst/go-writer"
	"log"
)

func main() {

	architecture_reader_uri := flag.String("architecture-reader-uri", "repo:///usr/local/data/sfomuseum-data-architecture", "")

	exhibition_reader_uri := flag.String("exhibitions-reader-uri", "repo:///usr/local/data/sfomuseum-data-exhibition", "")
	exhibition_writer_uri := flag.String("exhibitions-writer-uri", "", "If empty, the value of the -exhibition-reader-uri flag will be used.")

	exhibition_id := flag.Int64("exhibition-id", 0, "The SFO Museum exhibition ID to supersede")
	parent_id := flag.Int64("parent-id", 0, "The SFO Museum parent ID of the new exhibition")

	flag.Parse()

	ctx := context.Background()

	if *exhibition_writer_uri == "" {
		*exhibition_writer_uri = *exhibition_reader_uri
	}

	arch_r, err := reader.NewReader(ctx, *architecture_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create architecture reader, %v", err)
	}

	exh_r, err := reader.NewReader(ctx, *exhibition_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create exhibition reader, %v", err)
	}

	exh_wr, err := writer.NewWriter(ctx, *exhibition_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create exhibition writer, %v", err)
	}

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

		new_id, err := id_provider.NewID()

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

		_, err = sfom_writer.WriteFeatureBytes(ctx, exh_wr, new_exh)

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

		_, err = sfom_writer.WriteFeatureBytes(ctx, exh_wr, exh_f)

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
}
