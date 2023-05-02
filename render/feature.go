package render

import (
	"fmt"
)

type Feature struct {
	Id           int64
	ParentId     int64
	Name         string
	Inception    string
	Cessation    string
	Supersedes   []int64
	SupersededBy []int64
}

func (f *Feature) String() string {
	return fmt.Sprintf("%s\n%d (%d)\ninception: %s cessation: %s\nsupersedes %v\nsuperseded by: %v", f.Name, f.Id, f.ParentId, f.Inception, f.Cessation, f.Supersedes, f.SupersededBy)
}

var featureHash = func(f *Feature) string {
	return f.String()
}
