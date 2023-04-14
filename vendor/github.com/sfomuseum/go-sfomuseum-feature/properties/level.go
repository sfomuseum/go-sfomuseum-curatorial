package properties

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// Level returns the `sfo:level` property for 'body'.
func Level(body []byte) (int, error) {

	rsp := gjson.GetBytes(body, "properties.sfo:level")

	if !rsp.Exists() {
		return -1, fmt.Errorf("Missing sfo:level property")
	}

	return int(rsp.Int()), nil
}
