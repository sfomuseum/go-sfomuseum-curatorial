package publicart

import (
	"fmt"
)

type PublicArtWork struct {
	WhosOnFirstId int64  `json:"wof:id"`
	Name          string `json:"wof:name"`
	SFOMuseumId   int64  `json:"sfomuseum:object_id"`
	MapId         string `json:"sfomuseum:map_id"`

	// To do: is current stuff
}

func (w *PublicArtWork) String() string {
	return fmt.Sprintf("%d %s (%s)", w.WhosOnFirstId, w.Name, w.MapId)
}
