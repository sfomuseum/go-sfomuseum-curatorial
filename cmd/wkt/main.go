// Generate a CSV file with WOF ID, FileMaker ID and well-known-text (WKT) geometries for galleries and public art works marked "mz:is_current=1"
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/paulmach/orb/encoding/wkt"
	"github.com/sfomuseum/go-csvdict/v2"
	sfom_properties "github.com/sfomuseum/go-sfomuseum-feature/properties"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func main() {

	publicart_iterator_uri := flag.String("publicart-iterator-uri", "repo://?include=properties.mz:is_current=1", "")
	publicart_iterator_source := flag.String("publicart-iterator-source", "/usr/local/data/sfomuseum-data-publicart", "")

	galleries_iterator_uri := flag.String("galleries-iterator-uri", "repo://?include=properties.mz:is_current=1&include=properties.sfomuseum:placetype=gallery", "")
	galleries_iterator_source := flag.String("galleries-iterator-source", "/usr/local/data/sfomuseum-data-architecture", "")

	flag.Parse()

	var csv_wr *csvdict.Writer

	wr := os.Stdout

	mu := new(sync.RWMutex)

	data_sources := map[string]string{
		*publicart_iterator_uri: *publicart_iterator_source,
		*galleries_iterator_uri: *galleries_iterator_source,
	}

	derive_iterator_callback := func(source string) emitter.EmitterCallbackFunc {

		repo := filepath.Base(source)

		iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

			body, err := io.ReadAll(r)

			if err != nil {
				return fmt.Errorf("Failed to read %s, %w", path, err)
			}

			id, err := properties.Id(body)

			if err != nil {
				return fmt.Errorf("Failed to derive ID for %s, %w", path, err)
			}

			str_id := strconv.FormatInt(id, 10)

			pt, err := sfom_properties.Placetype(body)

			if err != nil {
				return fmt.Errorf("Failed to derive placetype for %s, %w", path, err)
			}

			var fmid_path string

			switch repo {
			case "sfomuseum-data-publicart":
				fmid_path = "properties.sfomuseum:object_id"
			case "sfomuseum-data-architecture":
				fmid_path = "properties.sfomuseum:gallery_id"
			default:
				return fmt.Errorf("Failed to derive FM path for %s", repo)
			}

			fmid_rsp := gjson.GetBytes(body, fmid_path)

			if !fmid_rsp.Exists() {
				return fmt.Errorf("Missing FM ID path %s for %s", fmid_path, path)
			}

			str_fmid := fmid_rsp.String()

			geojson_geom, err := geometry.Geometry(body)

			if err != nil {
				return fmt.Errorf("Failed to derive geometry for %s, %w", path, err)
			}

			orb_geom := geojson_geom.Geometry()

			wkt_geom := wkt.MarshalString(orb_geom)

			out := map[string]string{
				"wof_id":       str_id,
				"filemaker_id": str_fmid,
				"wkt":          wkt_geom,
				"placetype":    pt,
			}

			mu.Lock()
			defer mu.Unlock()

			if csv_wr == nil {

				w, err := csvdict.NewWriter(wr)

				if err != nil {
					return fmt.Errorf("Failed to create CSV writer, %w", err)
				}

				csv_wr = w
			}

			err = csv_wr.WriteRow(out)

			if err != nil {
				return fmt.Errorf("Failed to write CSV row for %s, %w", path, err)
			}

			return nil
		}

		return iter_cb
	}

	ctx := context.Background()
	wg := new(sync.WaitGroup)

	for iterator_uri, iterator_source := range data_sources {

		wg.Add(1)

		go func(iterator_uri string, iterator_source string) {

			defer wg.Done()

			iter_cb := derive_iterator_callback(iterator_source)

			iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

			if err != nil {
				log.Fatalf("Failed to create iterator for '%s', %v", iterator_uri, err)
			}

			err = iter.IterateURIs(ctx, iterator_source)

			if err != nil {
				log.Fatalf("Failed to iterate sources for '%s', %v", iterator_uri, err)
			}

		}(iterator_uri, iterator_source)
	}

	wg.Wait()

	csv_wr.Flush()
}
