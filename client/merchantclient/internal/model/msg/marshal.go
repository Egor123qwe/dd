package msg

import (
	"encoding/json"
)

func MarshalErr(t string, meta Meta, err error) []byte {
	meta.Err = &Err{
		Code:    "err",
		Message: err.Error(),
	}
	meta.Status = string(Error)

	return Marshal(t, meta, struct{}{})
}

func Marshal(t string, meta Meta, content any) []byte {
	m := MSG{
		Type: t,
		Meta: meta,
	}

	m.Content, _ = json.Marshal(content)
	result, _ := json.Marshal(m)

	return result
}
