package collection

import (
	"context"
	"fmt"

	"github.com/sfomuseum/go-sfomuseum-curatorial"
)

type Object struct {
	WhosOnFirstId   int64  `json:"wof:id"`
	Name            string `json:"wof:name"`
	SFOMuseumId     int64  `json:"sfomuseum:object_id"`
	AccessionNumber string `json:"sfomuseum:accession_number"`
	CallNumber      string `json:"sfomuseum:callnumber,omitempty"`
	IsCurrent       int64  `json:"mz:is_current"`
}

func (w *Object) String() string {
	return fmt.Sprintf("\"%s\"  %s %d (%d)", w.Name, w.AccessionNumber, w.WhosOnFirstId, w.SFOMuseumId)
}

// Return the current Object matching 'code'. Multiple matches throw an error.
func FindCurrentObject(ctx context.Context, code string) (*Object, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindCurrentObjectWithLookup(ctx, lookup, code)
}

// Return the current Object matching 'code' with a custom curatorial.Lookup instance. Multiple matches throw an error.
func FindCurrentObjectWithLookup(ctx context.Context, lookup curatorial.Lookup, code string) (*Object, error) {

	current, err := FindObjectsCurrentWithLookup(ctx, lookup, code)

	if err != nil {
		return nil, err
	}

	switch len(current) {
	case 0:
		return nil, NotFound{code}
	case 1:
		return current[0], nil
	default:
		return nil, MultipleCandidates{code}
	}

}

// Returns all Object instances matching 'code' that are marked as current.
func FindObjectsCurrent(ctx context.Context, code string) ([]*Object, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindObjectsCurrentWithLookup(ctx, lookup, code)
}

// Returns all Object instances matching 'code' that are marked as current with a custom curatorial.Lookup instance.
func FindObjectsCurrentWithLookup(ctx context.Context, lookup curatorial.Lookup, code string) ([]*Object, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, NotFound{code}
	}

	current := make([]*Object, 0)

	for _, r := range rsp {

		g := r.(*Object)

		// if g.IsCurrent == 0 {
		if g.IsCurrent != 1 {
			continue
		}

		current = append(current, g)
	}

	return current, nil
}
