package galleries

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

// CompileGalleriesData will generate a list of `Gallery` struct to be used as the source data for an `SFOMuseumLookup` instance.
// The list of gate are compiled by iterating over one or more source. `iterator_uri` is a valid `whosonfirst/go-whosonfirst-iterate` URI
// and `iterator_sources` are one more (iterator) URIs to process.
func CompileGalleriesData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Gallery, error) {

	lookup := make([]*Gallery, 0)
	mu := new(sync.RWMutex)

	iter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		if strings.HasSuffix(path, "~") {
			return nil
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
			return fmt.Errorf("Failed load feature from %s, %w", path, err)
		}

		wof_id, err := properties.Id(body)

		if err != nil {
			return fmt.Errorf("Failed to derive ID for %s, %w", path, err)
		}

		wof_name, err := properties.Name(body)

		if err != nil {
			return fmt.Errorf("Failed to derive name for %s, %w", path, err)
		}

		fl, err := properties.IsCurrent(body)

		if err != nil {
			return fmt.Errorf("Failed to determine is current for %s, %v", path, err)
		}

		sfomid_rsp := gjson.GetBytes(body, "properties.sfomuseum:gallery_id")

		if !sfomid_rsp.Exists() {
			return fmt.Errorf("Missing sfomuseum:gallery_id property (%s)", path)
		}

		mapid_rsp := gjson.GetBytes(body, "properties.sfomuseum:map_id")
		inception_rsp := gjson.GetBytes(body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(body, "properties.edtf:cessation")

		g := &Gallery{
			WhosOnFirstId: wof_id,
			SFOMuseumId:   sfomid_rsp.Int(),
			MapId:         mapid_rsp.String(),
			Name:          wof_name,
			Inception:     inception_rsp.String(),
			Cessation:     cessation_rsp.String(),
			IsCurrent:     fl.Flag(),
		}

		mu.Lock()
		lookup = append(lookup, g)
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
