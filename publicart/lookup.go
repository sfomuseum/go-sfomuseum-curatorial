package publicart

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sfomuseum/go-sfomuseum-curatorial"
	"github.com/sfomuseum/go-sfomuseum-curatorial/data"
)

var lookup_table *sync.Map
var lookup_idx int64

var lookup_init sync.Once
var lookup_init_err error

type PublicArtLookupFunc func(context.Context)

type PublicArtLookup struct {
	curatorial.Lookup
}

func init() {
	ctx := context.Background()
	curatorial.RegisterLookup(ctx, "publicart", NewLookup)
	lookup_idx = int64(0)
}

// NewLookup will return an `curatorial.Lookup` instance. By default the lookup table is derived from precompiled (embedded) data in `data/publicart.json`
// by passing in `publicart://` as the URI. It is also possible to create a new lookup table with the following URI options:
//
//	`publicart://github`
//
// This will cause the lookup table to be derived from the data stored at https://raw.githubusercontent.com/sfomuseum/go-sfomuseum-curatorial/main/data/publicart.json. This might be desirable if there have been updates to the underlying data that are not reflected in the locally installed package's pre-compiled data.
//
//	`publicart://iterator?uri={URI}&source={SOURCE}`
//
// This will cause the lookup table to be derived, at runtime, from data emitted by a `whosonfirst/go-whosonfirst-iterate` instance. `{URI}` should be a valid `whosonfirst/go-whosonfirst-iterate/iterator` URI and `{SOURCE}` is one or more URIs for the iterator to process.
func NewLookup(ctx context.Context, uri string) (curatorial.Lookup, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	// Reminder: u.Scheme is used by the curatorial.Lookup constructor

	switch u.Host {
	case "iterator":

		q := u.Query()

		iterator_uri := q.Get("uri")
		iterator_sources := q["source"]

		return NewLookupFromIterator(ctx, iterator_uri, iterator_sources...)

	case "github":

		data_url := "https://raw.githubusercontent.com/sfomuseum/go-sfomuseum-curatorial/main/data/publicart.json"
		rsp, err := http.Get(data_url)

		if err != nil {
			return nil, fmt.Errorf("Failed to load remote data from Github, %w", err)
		}

		lookup_func := NewLookupFuncWithReader(ctx, rsp.Body)
		return NewLookupWithLookupFunc(ctx, lookup_func)

	default:

		fs := data.FS
		fh, err := fs.Open("publicart.json")

		if err != nil {
			return nil, fmt.Errorf("Failed to load local precompiled data, %w", err)
		}

		lookup_func := NewLookupFuncWithReader(ctx, fh)
		return NewLookupWithLookupFunc(ctx, lookup_func)
	}
}

// NewLookup will return an `PublicArtLookupFunc` function instance that, when invoked, will populate an `curatorial.Lookup` instance with data stored in `r`.
// `r` will be closed when the `PublicArtLookupFunc` function instance is invoked.
// It is assumed that the data in `r` will be formatted in the same way as the procompiled (embedded) data stored in `data/sfomuseum.json`.
func NewLookupFuncWithReader(ctx context.Context, r io.ReadCloser) PublicArtLookupFunc {

	defer r.Close()

	var publicart_list []*PublicArtWork

	dec := json.NewDecoder(r)
	err := dec.Decode(&publicart_list)

	if err != nil {

		lookup_func := func(ctx context.Context) {
			lookup_init_err = err
		}

		return lookup_func
	}

	return NewLookupFuncWithPublicArtWorks(ctx, publicart_list)
}

// NewLookup will return an `PublicArtLookupFunc` function instance that, when invoked, will populate an `curatorial.Lookup` instance with data stored in `publicart_list`.
func NewLookupFuncWithPublicArtWorks(ctx context.Context, publicart_list []*PublicArtWork) PublicArtLookupFunc {

	lookup_func := func(ctx context.Context) {

		table := new(sync.Map)

		for _, data := range publicart_list {

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			appendData(ctx, table, data)
		}

		lookup_table = table
	}

	return lookup_func
}

// NewLookupWithLookupFunc will return an `curatorial.Lookup` instance derived by data compiled using `lookup_func`.
func NewLookupWithLookupFunc(ctx context.Context, lookup_func PublicArtLookupFunc) (curatorial.Lookup, error) {

	fn := func() {
		lookup_func(ctx)
	}

	lookup_init.Do(fn)

	if lookup_init_err != nil {
		return nil, lookup_init_err
	}

	l := PublicArtLookup{}
	return &l, nil
}

func NewLookupFromIterator(ctx context.Context, iterator_uri string, iterator_sources ...string) (curatorial.Lookup, error) {

	publicart_list, err := CompilePublicArtWorksData(ctx, iterator_uri, iterator_sources...)

	if err != nil {
		return nil, fmt.Errorf("Failed to compile public art work data, %w", err)
	}

	lookup_func := NewLookupFuncWithPublicArtWorks(ctx, publicart_list)
	return NewLookupWithLookupFunc(ctx, lookup_func)
}

func (l *PublicArtLookup) Find(ctx context.Context, code string) ([]interface{}, error) {

	pointers, ok := lookup_table.Load(code)

	if !ok {
		return nil, NotFound{code}
	}

	candidates := make([]interface{}, 0)

	for _, p := range pointers.([]string) {

		if !strings.HasPrefix(p, "pointer:") {
			return nil, fmt.Errorf("Invalid pointer, %s", p)
		}

		row, ok := lookup_table.Load(p)

		if !ok {
			return nil, fmt.Errorf("Invalid pointer, %s", p)
		}

		candidates = append(candidates, row.(*PublicArtWork))
	}

	return candidates, nil
}

func (l *PublicArtLookup) Append(ctx context.Context, data interface{}) error {
	return appendData(ctx, lookup_table, data.(*PublicArtWork))
}

func appendData(ctx context.Context, table *sync.Map, data *PublicArtWork) error {

	idx := atomic.AddInt64(&lookup_idx, 1)

	pointer := fmt.Sprintf("pointer:%d", idx)
	table.Store(pointer, data)

	str_wofid := strconv.FormatInt(data.WhosOnFirstId, 10)
	str_sfomid := strconv.FormatInt(data.SFOMuseumId, 10)

	possible_codes := []string{
		str_wofid,
		str_sfomid,
		fmt.Sprintf("wof:id=%s", str_wofid),
		fmt.Sprintf("sfomuseum:object_id=%s", str_sfomid),
	}

	if data.MapId != "" {
		possible_codes = append(possible_codes, data.MapId)
		possible_codes = append(possible_codes, fmt.Sprintf("sfomuseum:map_id=%s", data.MapId))
	}

	for _, code := range possible_codes {

		if code == "" {
			continue
		}

		pointers := make([]string, 0)
		has_pointer := false

		others, ok := table.Load(code)

		if ok {

			pointers = others.([]string)
		}

		for _, dupe := range pointers {

			if dupe == pointer {
				has_pointer = true
				break
			}
		}

		if has_pointer {
			continue
		}

		pointers = append(pointers, pointer)
		table.Store(code, pointers)
	}

	return nil
}
