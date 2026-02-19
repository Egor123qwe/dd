package parser

import (
	"encoding/json"
)

func ArrayToJSON[T any](arr []T) []string {
	var result []string

	for _, e := range arr {
		portJSON, err := json.Marshal(e)
		if err != nil {
			continue
		}

		result = append(result, string(portJSON))
	}

	return result
}
