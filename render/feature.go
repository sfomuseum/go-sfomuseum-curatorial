package render

import (
	"fmt"
)

type Feature struct {
	Id           int64
	Parent       *Feature
	ParentId     int64
	Name         string
	Inception    string
	Cessation    string
	Deprecated   string
	Supersedes   []int64
	SupersededBy []int64
}

func (f *Feature) String() string {

	name := f.Name

	if f.Deprecated != "" {
		name = fmt.Sprintf("%s\nDEPRECATED (%s)", name, f.Deprecated)
	}

	parent_label := ""

	if f.Parent != nil {
		parent_label = fmt.Sprintf("\n%s (%s - %s)", f.Parent.Name, f.Parent.Inception, f.Parent.Cessation)
	}

	return fmt.Sprintf("%s\n%d\n\nparent: %d%s\n\ninception: %s\ncessation: %s\n\nsupersedes %v\nsuperseded by: %v", name, f.Id, f.ParentId, parent_label, f.Inception, f.Cessation, f.Supersedes, f.SupersededBy)
}

var featureHash = func(f *Feature) string {
	return f.String()
}
