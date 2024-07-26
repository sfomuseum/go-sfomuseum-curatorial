package exhibitions

import (
	"context"
	"fmt"

	"github.com/sfomuseum/go-sfomuseum-curatorial"
)

type Exhibition struct {
	WhosOnFirstId  int64  `json:"wof:id"`
	Name           string `json:"wof:name"`
	SFOMuseumId    int64  `json:"sfomuseum:exhibition_id"`
	SFOMuseumWWWId int64  `json:"sfomuseum_www:exhibition_id"`
	IsCurrent      int64  `json:"mz:is_current"`

	// To do: is current stuff
	// To do (maybe): galleries
}

func (w *Exhibition) String() string {
	return fmt.Sprintf("%d %s FM: %d WWW: %d", w.WhosOnFirstId, w.Name, w.SFOMuseumId, w.SFOMuseumWWWId)
}

// Return the current Exhibition matching 'code'. Multiple matches throw an error.
func FindCurrentExhibition(ctx context.Context, code string) (*Exhibition, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindCurrentExhibitionWithLookup(ctx, lookup, code)
}

// Return the current Exhibition matching 'code' with a custom curatorial.Lookup instance. Multiple matches throw an error.
func FindCurrentExhibitionWithLookup(ctx context.Context, lookup curatorial.Lookup, code string) (*Exhibition, error) {

	current, err := FindExhibitionsCurrentWithLookup(ctx, lookup, code)

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

// Returns all Exhibition instances matching 'code' that are marked as current.
func FindExhibitionsCurrent(ctx context.Context, code string) ([]*Exhibition, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindExhibitionsCurrentWithLookup(ctx, lookup, code)
}

// Returns all Exhibition instances matching 'code' that are marked as current with a custom curatorial.Lookup instance.
func FindExhibitionsCurrentWithLookup(ctx context.Context, lookup curatorial.Lookup, code string) ([]*Exhibition, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, NotFound{code}
	}

	current := make([]*Exhibition, 0)

	for _, r := range rsp {

		g := r.(*Exhibition)

		// if g.IsCurrent == 0 {
		if g.IsCurrent != 1 {
			continue
		}

		current = append(current, g)
	}

	return current, nil
}
