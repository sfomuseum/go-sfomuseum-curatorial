package main

import (
	_ "github.com/sfomuseum/go-sfomuseum-curatorial/collection"
	_ "github.com/sfomuseum/go-sfomuseum-curatorial/exhibitions"
	_ "github.com/sfomuseum/go-sfomuseum-curatorial/publicart"
)

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/sfomuseum/go-sfomuseum-curatorial"
)

func main() {

	schemes := curatorial.LookupSchemes()

	lookup_uri_desc := fmt.Sprintf("Valid options are: %s", strings.Join(schemes, ", "))
	lookup_uri := flag.String("lookup-uri", "", lookup_uri_desc)

	flag.Parse()

	ctx := context.Background()

	lookup, err := curatorial.NewLookup(ctx, *lookup_uri)

	if err != nil {
		log.Fatalf("Failed to create new lookup for %s, %v", *lookup_uri, err)
	}

	for _, code := range flag.Args() {

		results, err := lookup.Find(ctx, code)

		if err != nil {
			log.Fatal(err)
		}

		for _, a := range results {
			fmt.Println(a)
		}
	}
}
