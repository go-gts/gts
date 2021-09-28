package gts

import (
	"encoding/json"
	"strings"
)

func jsonify(v interface{}) string {
	p, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(p)
}

func multiLineString(ss ...string) string {
	return strings.Join(ss, "\n")
}
