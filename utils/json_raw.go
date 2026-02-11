package utils

import (
	"encoding/json"
	"errors"
)

// JSONRaw preserves raw JSON bytes in request/response payloads.
type JSONRaw []byte

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSONRaw) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return errors.New("invalid JSON")
	}
	*j = append((*j)[0:0], data...)
	return nil
}
