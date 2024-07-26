package galleries

import (
	"context"
	"fmt"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/whosonfirst/go-reader"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
)

// Derive a MultiPoint geoemtry for one or more gallery IDs.
func GeometryForGalleryIDs(ctx context.Context, r reader.Reader, gallery_ids ...int64) (orb.Geometry, error) {

	count_galleries := len(gallery_ids)

	switch count_galleries {
	case 0:
		return nil, fmt.Errorf("Please Null Terminal me")
	case 1:

		points, err := multipoints(ctx, r, gallery_ids[0])

		if err != nil {
			return nil, err
		}

		return orb.MultiPoint(points), nil

	default:

		points := make([]orb.Point, 0)

		for _, wofid := range gallery_ids {

			mp, err := multipoints(ctx, r, wofid)

			if err != nil {
				return nil, err
			}

			for _, p := range mp {
				points = append(points, p)
			}
		}

		return orb.MultiPoint(points), nil
	}
}

// Derive a MultiPoint geometry for a Who's On First (gallery) ID.
// In the future we expect that all galleries will be defined as MultiPolygons but
// today they are not.
func multipoints(ctx context.Context, r reader.Reader, wofid int64) (orb.MultiPoint, error) {

	body, err := wof_reader.LoadBytes(ctx, r, wofid)

	if err != nil {
		return nil, err
	}

	f, err := geojson.UnmarshalFeature(body)

	if err != nil {
		return nil, err
	}

	geom := f.Geometry

	switch geom.GeoJSONType() {
	case "Point":

		points := []orb.Point{
			geom.(orb.Point),
		}

		return orb.MultiPoint(points), nil

	case "MultiPoint":
		return geom.(orb.MultiPoint), nil
	case "MultiPolygon":

		points := make([]orb.Point, 0)

		for _, poly := range geom.(orb.MultiPolygon) {
			pt, _ := planar.CentroidArea(poly)
			points = append(points, pt)
		}

		return orb.MultiPoint(points), nil

	case "Polygon":

		pt, _ := planar.CentroidArea(geom)
		points := []orb.Point{pt}

		return orb.MultiPoint(points), nil

	default:
		return nil, fmt.Errorf("Weirdo geometry type for gallery %d, %s", wofid, geom.GeoJSONType())
	}

}
