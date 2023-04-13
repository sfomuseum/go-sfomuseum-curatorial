package exhibitions

import (
	"context"
	"fmt"
	"io"
	_ "log"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func CompileExhibitionsData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Exhibition, error) {

	lookup := make([]*Exhibition, 0)
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

		wofid_rsp := gjson.GetBytes(body, "properties.wof:id")
		sfomid_rsp := gjson.GetBytes(body, "properties.sfomuseum:exhibition_id")

		if !wofid_rsp.Exists() {
			return fmt.Errorf("'%s' is missing wof:id property", path)
		}

		if !sfomid_rsp.Exists() {
			return fmt.Errorf("'%s' is missing sfomuseum:exhibition_id property", path)
		}

		name_rsp := gjson.GetBytes(body, "properties.wof:name")

		is_current, err := properties.IsCurrent(body)

		if err != nil {
			return fmt.Errorf("Failed to derive is current for %s, %w", path, err)
		}

		w := &Exhibition{
			WhosOnFirstId: wofid_rsp.Int(),
			SFOMuseumId:   sfomid_rsp.Int(),
			Name:          name_rsp.String(),
			IsCurrent:     is_current.Flag(),
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
