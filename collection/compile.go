package collection

import (
	"context"
	"fmt"
	"io"
	_ "log"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func CompileCollectionData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Object, error) {

	lookup := make([]*Object, 0)

	iter, err := iterate.NewIterator(ctx, iterator_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create iterator, %w", err)
	}

	for rec, err := range iter.Iterate(ctx, iterator_sources...) {
		
		if err != nil {
			return nil, fmt.Errorf("Failed to iterate sources, %w", err)
		}

		defer rec.Body.Close()
		
		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		_, uri_args, err := uri.ParseURI(rec.Path)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse %s, %w", rec.Path, err)
		}

		if uri_args.IsAlternate {
			continue
		}

		body, err := io.ReadAll(rec.Body)

		if err != nil {
			return nil, fmt.Errorf("Failed to read '%s', %w", rec.Path, err)
		}

		wof_id, err := properties.Id(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive wof:id for %s, %w", rec.Path, err)
		}

		wof_name, err := properties.Name(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive wof:name for %s, %w", rec.Path, err)
		}

		is_current, err := properties.IsCurrent(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive is current for %s, %w", rec.Path, err)
		}

		sfomid_rsp := gjson.GetBytes(body, "properties.sfomuseum:object_id")

		if !sfomid_rsp.Exists() {
			return nil, fmt.Errorf("'%s' is missing sfomuseum:object_id property", rec.Path)
		}

		accno_rsp := gjson.GetBytes(body, "properties.sfomuseum:accession_number")

		if !sfomid_rsp.Exists() {
			return nil, fmt.Errorf("'%s' is missing sfomuseum:accession_number property", rec.Path)
		}

		w := &Object{
			WhosOnFirstId:   wof_id,
			SFOMuseumId:     sfomid_rsp.Int(),
			AccessionNumber: accno_rsp.String(),
			Name:            wof_name,
			IsCurrent:       is_current.Flag(),
		}

		callno_rsp := gjson.GetBytes(body, "properties.sfomuseum:callnumber")

		if callno_rsp.Exists() && callno_rsp.String() != "" {
			w.CallNumber = callno_rsp.String()
		}

		lookup = append(lookup, w)
	}

	return lookup, nil
}
