package publicart

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-curatorial"
)

type PublicArtWork struct {
	WhosOnFirstId int64  `json:"wof:id"`
	Name          string `json:"wof:name"`
	SFOMuseumId   int64  `json:"sfomuseum:object_id"`
	MapId         string `json:"sfomuseum:map_id"`
	IsCurrent     int64  `json:"mz:is_current"`
}

func (w *PublicArtWork) String() string {
	return fmt.Sprintf("\"%s\" %d (%d) (%s) Is current: %d", w.Name, w.WhosOnFirstId, w.SFOMuseumId, w.MapId, w.IsCurrent)
}

// Return the current PublicArtWork matching 'code'. Multiple matches throw an error.
func FindCurrentPublicArtWork(ctx context.Context, code string) (*PublicArtWork, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindCurrentPublicArtWorkWithLookup(ctx, lookup, code)
}

// Return the current PublicArtWork matching 'code' with a custom curatorial.Lookup instance. Multiple matches throw an error.
func FindCurrentPublicArtWorkWithLookup(ctx context.Context, lookup curatorial.Lookup, code string) (*PublicArtWork, error) {

	current, err := FindPublicArtWorksCurrentWithLookup(ctx, lookup, code)

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

// Returns all PublicArtWork instances matching 'code' that are marked as current.
func FindPublicArtWorksCurrent(ctx context.Context, code string) ([]*PublicArtWork, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindPublicArtWorksCurrentWithLookup(ctx, lookup, code)
}

// Returns all PublicArtWork instances matching 'code' that are marked as current with a custom curatorial.Lookup instance.
func FindPublicArtWorksCurrentWithLookup(ctx context.Context, lookup curatorial.Lookup, code string) ([]*PublicArtWork, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, NotFound{code}
	}

	current := make([]*PublicArtWork, 0)

	for _, r := range rsp {

		g := r.(*PublicArtWork)

		// if g.IsCurrent == 0 {
		if g.IsCurrent != 1 {
			continue
		}

		current = append(current, g)
	}

	return current, nil
}
