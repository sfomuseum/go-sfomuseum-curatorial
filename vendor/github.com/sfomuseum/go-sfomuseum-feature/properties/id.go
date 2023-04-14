package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

// Id returns the `sfomuseum:*` identifier for 'body' based on its placetype.
func Id(body []byte) (int64, error) {

	pt, err := Placetype(body)

	if err != nil {
		return 0, fmt.Errorf("Failed to derive placetype, %w", err)
	}

	var k string

	switch pt {
	case "aircraft", "airline", "airport":
		k = fmt.Sprintf("sfomuseum:%s_id", pt)
	case "exhibition", "gallery":
		k = fmt.Sprintf("sfomuseum:%s_id", pt)
	case "object", "publicart":
		k = "sfomuseum:object_id"
	default:
		return 0, fmt.Errorf("Unrecognized placetype, %s", pt)
	}

	path := fmt.Sprintf("properties.%s", k)
	rsp := gjson.GetBytes(body, path)

	if !rsp.Exists() {
		return 0, fmt.Errorf("Derived ID path (%s) does not exist", k)
	}

	return rsp.Int(), nil
}
