package properties

import (
	"github.com/tidwall/gjson"
)

// PostSecurity returns the `sfomuseum:post_security` identifier for 'body' based or '-1' if it can not be determine.
func PostSecurity(body []byte) int {

	rsp := gjson.GetBytes(body, "properties.sfomuseum:post_security")

	if !rsp.Exists() {
		return -1
	}

	return int(rsp.Int())
}
