package publicart

import (
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-iterate/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	_ "log"
	"sync"
)

func CompilePublicArtWorksData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*PublicArtWork, error) {

	lookup := make([]*PublicArtWork, 0)
	mu := new(sync.RWMutex)

	iter_cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		path, err := emitter.PathForContext(ctx)

		if err != nil {
			return fmt.Errorf("Failed to derive path from context, %w", err)
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
		sfomid_rsp := gjson.GetBytes(body, "properties.sfomuseum:object_id")

		if !wofid_rsp.Exists() {
			return fmt.Errorf("'%s' is missing wof:id property", path)
		}

		if !sfomid_rsp.Exists() {
			return fmt.Errorf("'%s' is missing sfomuseum:obhect_id property", path)
		}

		name_rsp := gjson.GetBytes(body, "properties.wof:name")
		mapid_rsp := gjson.GetBytes(body, "properties.sfomuseum:map_id")

		w := &PublicArtWork{
			WhosOnFirstId: wofid_rsp.Int(),
			SFOMuseumId:   sfomid_rsp.Int(),
			Name:          name_rsp.String(),
			MapId:         mapid_rsp.String(),
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
