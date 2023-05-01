package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"bytes"
	"bufio"
	"image/png"

	"github.com/goccy/go-graphviz"	
	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

type Feature struct {
	Id int64
	ParentId int64
	Name string
	Inception string
	Cessation string
	Supersedes []int64
	SupersededBy []int64
}

func (f *Feature) String() string {
	return fmt.Sprintf("%s\n%d (%d)\ninception: %s cessation: %s\nsupersedes %v\nsuperseded by: %v", f.Name, f.Id, f.ParentId, f.Inception, f.Cessation, f.Supersedes, f.SupersededBy)
}

func main() {

	iterator_uri := flag.String("iterator-uri", "", "...")
	iterator_source := flag.String("iterator-source", "", "...")
	
	reader_uri := flag.String("reader-uri", "", "")
	parent_reader_uri := flag.String("parent-reader-uri", "", "")

	flag.Parse()

	ctx := context.Background()

	feature_r, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new reader, %v", err)
	}

	parent_r, err := reader.NewReader(ctx, *parent_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new parent reader, %v", err)
	}

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
	
	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse '%s', %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		f, err := deriveFeatureWithBody(ctx, body, id)

		if err != nil {
			return fmt.Errorf("Failed to derive feature for %s, %w", path, err)
		}
				
		g.AddVertex(f, feature_attrs...)

		for _, other_id := range f.Supersedes {
				
			other_f, err := deriveFeature(ctx, feature_r, other_id)

			if err != nil {
				return fmt.Errorf("Failed to derive feature for supersedes ID (%d) for %s, %w", other_id, path, err)
			}
			
			g.AddVertex(other_f, feature_attrs...)
			g.AddEdge(f.String(), other_f.String())
		}

		for _, other_id := range f.SupersededBy {

			other_f, err := deriveFeature(ctx, feature_r, other_id)

			if err != nil {
				return fmt.Errorf("Failed to derive feature for superseded_by ID (%d) for %s, %w", other_id, path, err)
			}
			
			g.AddVertex(other_f, feature_attrs...)
			g.AddEdge(f.String(), other_f.String())
		}

		// return nil
		
		parent_f, err := deriveFeature(ctx, parent_r, f.ParentId)
		
		if err != nil {
			return fmt.Errorf("Failed to load parent ID (%d) for %s, %w", f.ParentId, path, err)
		}
		
		g.AddVertex(parent_f, parent_attrs...)			
		g.AddEdge(f.String(), parent_f.String())

		
		for _, other_id := range parent_f.Supersedes {
				
			other_f, err := deriveFeature(ctx, parent_r, other_id)

			if err != nil {
				return fmt.Errorf("Failed to derive feature for parent supersedes ID (%d) for %s, %w", other_id, path, err)
			}
			
			g.AddVertex(other_f, parent_attrs...)
			g.AddEdge(parent_f.String(), other_f.String())
		}

		return nil
		
		for _, other_id := range parent_f.SupersededBy {

			other_f, err := deriveFeature(ctx, parent_r, other_id)

			if err != nil {
				return fmt.Errorf("Failed to derive feature for parent superseded_by ID (%d) for %s, %w", other_id, path, err)
			}
			
			g.AddVertex(other_f, parent_attrs...)
			g.AddEdge(other_f.String(), parent_f.String())
		}
		
		return nil
	}

	iter, err := iterator.NewIterator(ctx, *iterator_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to iterate source, %v", err)
	}

	// Dot stuff
	
	var buf bytes.Buffer
	buf_wr := bufio.NewWriter(&buf)

	err = draw.DOT(g, buf_wr)
	
	buf_wr.Flush()

	// Graphviz (image) stuff
	
	gv := graphviz.New()

	graph, err := graphviz.ParseBytes(buf.Bytes())

	if err != nil {
		log.Fatalf("Failed to parse graphviz data, %v", err)
	}

	im, err := gv.RenderImage(graph)

	if err != nil {
		log.Fatalf("Failed to render graphviz data, %v", err)
	}

	// Image (on disk) stuff
	
	out, err  := os.OpenFile("graph.png", os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalf("Failed to open image file, %v", err)
	}

	err = png.Encode(out, im)

	if err != nil {
		log.Fatalf("Failed to encode PNG image, %v", err)
	}

	err = out.Close()

	if err != nil {
		log.Fatalf("Failed to close PNG imge, %v", err)
	}
}

func deriveFeature(ctx context.Context, r reader.Reader, id int64) (*Feature, error) {

	body, err := wof_reader.LoadBytes(ctx, r, id)
	
	if err != nil {
		return nil, fmt.Errorf("Failed to load %d, %w", id, err)
	}
	
	return deriveFeatureWithBody(ctx, body, id)
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
	
	supersedes := properties.Supersedes(body)
	superseded_by := properties.SupersededBy(body)		
	
	f := &Feature{
		Id: id,
		Name: name,
		ParentId: parent_id,
		Inception: inception,
		Cessation: cessation,
		Supersedes: supersedes,
		SupersededBy: superseded_by,			
	}
	
	return f, nil
}

