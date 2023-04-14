package properties

import (
	"fmt"
	"github.com/tidwall/gjson"
)

// Placetype returns the value of the `sfomuseum:placetype` property for 'body'.
func Placetype(body []byte) (string, error) {

	rsp := gjson.GetBytes(body, "properties.sfomuseum:placetype")

	if !rsp.Exists() {
		return "", fmt.Errorf("Missing sfomuseum:placetype property")
	}

	placetype := rsp.String()

	return placetype, nil
}
