package render

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image/png"
	"io"
	"log"
	"sync/atomic"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
	"github.com/goccy/go-graphviz"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

type RenderOptions struct {
	FeatureReader   reader.Reader
	ParentReader    reader.Reader
	IteratorURI     string
	IteratorSources []string
}

func Render(ctx context.Context, opts *RenderOptions) (uint32, []byte, error) {

	featureHash := func(f *Feature) string {
		return f.String()
	}

	feature_attrs := []func(*graph.VertexProperties){
		graph.VertexAttribute("shape", "box"),
		graph.VertexAttribute("color", "black"),
		graph.VertexAttribute("decorate", "true"),
		graph.VertexAttribute("fontsize", "10"),
		graph.VertexAttribute("linelength", "150"),
		graph.VertexAttribute("margin", ".5"),
	}

	parent_attrs := []func(*graph.VertexProperties){
		graph.VertexAttribute("shape", "ellipse"),
		graph.VertexAttribute("color", "grey"),
		graph.VertexAttribute("decorate", "true"),
		graph.VertexAttribute("fontsize", "10"),
		graph.VertexAttribute("linelength", "150"),
		graph.VertexAttribute("margin", ".5"),
	}

	g := graph.New(featureHash, graph.Directed(), graph.Acyclic())

	count := uint32(0)

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse '%s', %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		/*
			body, err := io.ReadAll(r)

			if err != nil {
				return fmt.Errorf("Failed to read %s, %w", path, err)
			}
		*/

		f, err := deriveFeature(ctx, opts.FeatureReader, opts.ParentReader, id)

		if err != nil {
			return fmt.Errorf("Failed to derive feature for %s, %w", path, err)
		}

		g.AddVertex(f, feature_attrs...)

		atomic.AddUint32(&count, 1)

		for _, other_id := range f.Supersedes {

			other_f, err := deriveFeature(ctx, opts.FeatureReader, opts.ParentReader, other_id)

			if err != nil {
				log.Printf("Failed to derive feature for supersedes ID (%d) for %s, %w", other_id, path, err)
				continue
			}

			g.AddVertex(other_f, feature_attrs...)
			g.AddEdge(f.String(), other_f.String())
		}

		for _, other_id := range f.SupersededBy {

			other_f, err := deriveFeature(ctx, opts.FeatureReader, opts.ParentReader, other_id)

			if err != nil {
				log.Printf("Failed to derive feature for superseded_by ID (%d) for %s, %w", other_id, path, err)
				continue
			}

			g.AddVertex(other_f, feature_attrs...)
			g.AddEdge(f.String(), other_f.String())
		}

		return nil

		if f.ParentId == -1 {
			return nil
		}

		// return nil

		parent_f, err := deriveFeature(ctx, opts.ParentReader, opts.ParentReader, f.ParentId)

		if err != nil {
			return fmt.Errorf("Failed to load parent ID (%d) for %s, %w", f.ParentId, path, err)
		}

		g.AddVertex(parent_f, parent_attrs...)
		g.AddEdge(parent_f.String(), f.String())

		for _, other_id := range parent_f.Supersedes {

			other_f, err := deriveFeature(ctx, opts.ParentReader, opts.ParentReader, other_id)

			if err != nil {
				log.Printf("Failed to derive feature for parent supersedes ID (%d) for %s, %w", other_id, path, err)
				continue
			}

			g.AddVertex(other_f, parent_attrs...)
			g.AddEdge(other_f.String(), parent_f.String())
		}

		return nil

		for _, other_id := range parent_f.SupersededBy {

			other_f, err := deriveFeature(ctx, opts.ParentReader, opts.ParentReader, other_id)

			if err != nil {
				log.Printf("Failed to derive feature for parent superseded_by ID (%d) for %s, %w", other_id, path, err)
				continue
			}

			g.AddVertex(other_f, parent_attrs...)
			g.AddEdge(other_f.String(), parent_f.String())
		}

		return nil
	}

	iter, err := iterator.NewIterator(ctx, opts.IteratorURI, iter_cb)

	if err != nil {
		return 0, nil, fmt.Errorf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, opts.IteratorSources...)

	if err != nil {
		return 0, nil, fmt.Errorf("Failed to iterate source, %v", err)
	}

	// Dot stuff

	var buf bytes.Buffer
	buf_wr := bufio.NewWriter(&buf)

	err = draw.DOT(g, buf_wr)

	if err != nil {
		return 0, nil, fmt.Errorf("Failed to render graph as dot, %w", err)
	}

	buf_wr.Flush()

	return count, buf.Bytes(), nil
}

func Draw(ctx context.Context, body []byte, wr io.Writer) error {

	// Graphviz (image) stuff

	gv := graphviz.New()

	graph, err := graphviz.ParseBytes(body)

	if err != nil {
		return fmt.Errorf("Failed to parse graphviz data, %v", err)
	}

	im, err := gv.RenderImage(graph)

	if err != nil {
		return fmt.Errorf("Failed to render graphviz data, %v", err)
	}

	// Image (on disk) stuff

	err = png.Encode(wr, im)

	if err != nil {
		return fmt.Errorf("Failed to encode PNG image, %v", err)
	}

	return nil
}

func deriveFeature(ctx context.Context, r reader.Reader, parent_r reader.Reader, id int64) (*Feature, error) {

	body, err := wof_reader.LoadBytes(ctx, r, id)

	if err != nil {
		return nil, fmt.Errorf("Failed to load %d, %w", id, err)
	}

	f, err := deriveFeatureWithBody(ctx, body, id)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive feature for %d, %w", id, err)
	}

	if f.ParentId > -1 {

		parent_body, err := wof_reader.LoadBytes(ctx, parent_r, f.ParentId)

		if err != nil {
			return nil, fmt.Errorf("Failed to load parent record for %d, %w", f.ParentId, err)
		}

		parent_f, err := deriveFeatureWithBody(ctx, parent_body, f.ParentId)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive parent feature for %d, %w", f.ParentId, err)
		}

		f.Parent = parent_f
	}

	return f, nil
}

func deriveFeatureWithBody(ctx context.Context, body []byte, id int64) (*Feature, error) {

	name, err := properties.Name(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive name for %d, %w", id, err)
	}

	parent_id, err := properties.ParentId(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive parent ID for %d, %w", id, err)
	}

	inception := properties.Inception(body)
	cessation := properties.Cessation(body)
	deprecated := properties.Deprecated(body)

	supersedes := properties.Supersedes(body)
	superseded_by := properties.SupersededBy(body)

	f := &Feature{
		Id:           id,
		Name:         name,
		ParentId:     parent_id,
		Inception:    inception,
		Cessation:    cessation,
		Deprecated:   deprecated,
		Supersedes:   supersedes,
		SupersededBy: superseded_by,
	}

	return f, nil
}
