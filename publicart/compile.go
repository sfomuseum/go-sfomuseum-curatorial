package publicart

import (
	"context"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func CompilePublicArtWorksData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*PublicArtWork, error) {

	lookup := make([]*PublicArtWork, 0)

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

		wofid_rsp := gjson.GetBytes(body, "properties.wof:id")
		sfomid_rsp := gjson.GetBytes(body, "properties.sfomuseum:object_id")

		if !wofid_rsp.Exists() {
			return nil, fmt.Errorf("'%s' is missing wof:id property", rec.Path)
		}

		if !sfomid_rsp.Exists() {
			// slog.Warn("Record is missing sfomuseum:obhect_id property, skipping", "path", rec.Path)
			// return nil
			return nil, fmt.Errorf("'%s' is missing sfomuseum:obhect_id property", rec.Path)
		}

		name_rsp := gjson.GetBytes(body, "properties.wof:name")

		is_current, err := properties.IsCurrent(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive is current for %s, %w", rec.Path, err)
		}

		w := &PublicArtWork{
			WhosOnFirstId: wofid_rsp.Int(),
			SFOMuseumId:   sfomid_rsp.Int(),
			Name:          name_rsp.String(),
			IsCurrent:     is_current.Flag(),
		}

		mapid_rsp := gjson.GetBytes(body, "properties.sfomuseum:map_id")

		if mapid_rsp.String() != "" {
			w.MapId = mapid_rsp.String()
		}

		lookup = append(lookup, w)
	}

	return lookup, nil
}
