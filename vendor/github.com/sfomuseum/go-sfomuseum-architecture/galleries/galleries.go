// package galleries provides methods for working with boarding galleries at SFO.
package galleries

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfomuseum/go-edtf/cmp"
	"github.com/sfomuseum/go-sfomuseum-architecture"
)

// type Gallery is a struct representing a passenger gallery at SFO.
type Gallery struct {
	// The Who's On First ID associated with this gallery.
	WhosOnFirstId int64 `json:"wof:id"`
	// The SFO Museum ID associated with this gallery.
	SFOMuseumId int64 `json:"sfomuseum:id"`
	// The map label (ID) associated with this gallery.
	MapId string `json:"map_id"`
	// The name of this gallery.
	Name string `json:"wof:name"`
	// The (EDTF) inception date for the gallery
	Inception string `json:"edtf:inception"`
	// The (EDTF) cessation date for the gallery
	Cessation string `json:"edtf:cessation"`
	// A Who's On First "existential" (`KnownUnknownFlag`) flag signaling the gallery's status
	IsCurrent int64 `json:"mz:is_current"`
}

// String() will return the name of the gallery.
func (g *Gallery) String() string {
	return fmt.Sprintf("%d#%d %s %s-%s (%d)", g.WhosOnFirstId, g.SFOMuseumId, g.Name, g.Inception, g.Cessation, g.IsCurrent)
}

// Return the Gallery matching 'code' that was active for 'date'. Multiple matches throw an error.
func FindGalleryForDate(ctx context.Context, code string, date string) (*Gallery, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindGalleryForDateWithLookup(ctx, lookup, code, date)
}

// Return all the Galleries matching 'code' that were active for 'date'.
func FindAllGalleriesForDate(ctx context.Context, code string, date string) ([]*Gallery, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindAllGalleriesForDateWithLookup(ctx, lookup, code, date)
}

// Return the current Gallery matching 'code'. Multiple matches throw an error.
func FindCurrentGallery(ctx context.Context, code string) (*Gallery, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindCurrentGalleryWithLookup(ctx, lookup, code)
}

// Return the current Gallery matching 'code' with a custom architecture.Lookup instance. Multiple matches throw an error.
func FindCurrentGalleryWithLookup(ctx context.Context, lookup architecture.Lookup, code string) (*Gallery, error) {

	current, err := FindGalleriesCurrentWithLookup(ctx, lookup, code)

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

// Returns all Gallery instances matching 'code' that are marked as current.
func FindGalleriesCurrent(ctx context.Context, code string) ([]*Gallery, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindGalleriesCurrentWithLookup(ctx, lookup, code)
}

// Returns all Gallery instances matching 'code' that are marked as current with a custom architecture.Lookup instance.
func FindGalleriesCurrentWithLookup(ctx context.Context, lookup architecture.Lookup, code string) ([]*Gallery, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("Failed to find %s, %w", code, err)
	}

	current := make([]*Gallery, 0)

	for _, r := range rsp {

		g := r.(*Gallery)

		// if g.IsCurrent == 0 {
		if g.IsCurrent != 1 {
			continue
		}

		current = append(current, g)
	}

	return current, nil
}

// Return the Gallery matching 'code' that was active for 'date' using 'lookup'. Multiple matches throw an error.
func FindGalleryForDateWithLookup(ctx context.Context, lookup architecture.Lookup, code string, date string) (*Gallery, error) {

	galleries, err := FindAllGalleriesForDateWithLookup(ctx, lookup, code, date)

	if err != nil {
		return nil, err
	}

	switch len(galleries) {
	case 0:
		return nil, NotFound{code}
	case 1:
		return galleries[0], nil
	default:
		return nil, MultipleCandidates{code}
	}

}

// Return all the Gallerys matching 'code' that were active for 'date' using 'lookup'.
func FindAllGalleriesForDateWithLookup(ctx context.Context, lookup architecture.Lookup, code string, date string) ([]*Gallery, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("Failed to find gallerys for code, %w", err)
	}

	galleries := make([]*Gallery, 0)

	for _, r := range rsp {

		g := r.(*Gallery)

		inception := g.Inception
		cessation := g.Cessation

		is_between, err := cmp.IsBetween(date, inception, cessation)

		if err != nil {
			slog.Debug("Failed to determine whether gallery matches date conditions", "code", code, "date", date, "gallery", g.Name, "inception", inception, "cessation", cessation, "error", err)
			continue
		}

		if !is_between {
			slog.Debug("Gallery does not match date conditions", "code", code, "date", date, "gallery", g.Name, "inception", inception, "cessation", cessation)
			continue
		}

		slog.Debug("Gallery DOES match date conditions", "code", code, "date", date, "gallery", g.Name, "inception", inception, "cessation", cessation)
		galleries = append(galleries, g)
		break
	}

	slog.Debug("Return galleries", "code", code, "date", date, "count", len(galleries))
	return galleries, nil
}
