package json_field

import (
	"encoding/json"
)

type JsonField[T any] struct {
	// DANGEROUS !!! DO NOT CHANGE NAME (valuer.go field name dependency)
	Present bool

	// DANGEROUS !!! DO NOT CHANGE NAME (valuer.go field name dependency)
	Value *T
}

func (v *JsonField[T]) UnmarshalJSON(data []byte) error {
	v.Present = true

	if string(data) == "null" {
		return nil
	}

	var value T

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	v.Value = &value

	return nil
}

func (v *JsonField[T]) Convert() **T {
	if !v.Present {
		return nil
	}

	return &v.Value
}
