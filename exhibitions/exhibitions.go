package exhibitions

import (
	"fmt"
)

type Exhibition struct {
	WhosOnFirstId int64  `json:"wof:id"`
	Name          string `json:"wof:name"`
	SFOMuseumId   int64  `json:"sfomuseum:object_id"`

	// To do: is current stuff
	// To do (maybe): galleries
}

func (w *Exhibition) String() string {
	return fmt.Sprintf("%d %s", w.WhosOnFirstId, w.Name)
}
