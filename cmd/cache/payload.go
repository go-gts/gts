package cache

import (
	"encoding/json"
	"sort"
)

type Pair [2]interface{}

type Payload map[string]interface{}

func EncodePayload(payload map[string]interface{}) ([]byte, error) {
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	tpls := make([]Pair, len(keys))
	for i, key := range keys {
		tpls[i] = Pair{key, payload[key]}
	}
	return json.Marshal(tpls)
}
