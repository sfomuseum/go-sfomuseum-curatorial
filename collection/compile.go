package collection

import (
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	_ "log"
	"sync"
)

func CompileCollectionData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Object, error) {

	lookup := make([]*Object, 0)
	mu := new(sync.RWMutex)

	iter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		_, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse %s, %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read '%s', %w", path, err)
		}

		wof_id, err := properties.Id(body)

		if err != nil {
			return fmt.Errorf("Failed to derive wof:id for %s, %w", path, err)
		}

		wof_name, err := properties.Name(body)

		if err != nil {
			return fmt.Errorf("Failed to derive wof:name for %s, %w", path, err)
		}

		is_current, err := properties.IsCurrent(body)

		if err != nil {
			return fmt.Errorf("Failed to derive is current for %s, %w", path, err)
		}

		sfomid_rsp := gjson.GetBytes(body, "properties.sfomuseum:object_id")

		if !sfomid_rsp.Exists() {
			return fmt.Errorf("'%s' is missing sfomuseum:object_id property", path)
		}

		accno_rsp := gjson.GetBytes(body, "properties.sfomuseum:accession_number")

		if !sfomid_rsp.Exists() {
			return fmt.Errorf("'%s' is missing sfomuseum:accession_number property", path)
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

		mu.Lock()
		lookup = append(lookup, w)
		mu.Unlock()

		return nil
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		return nil, fmt.Errorf("Failed to create iterator, %w", err)
	}

	err = iter.IterateURIs(ctx, iterator_sources...)

	if err != nil {
		return nil, fmt.Errorf("Failed to iterate sources, %w", err)
	}

	return lookup, nil
}
