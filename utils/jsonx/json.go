package jsonx

import (
	"encoding/json"
	"errors"
	"fmt"
)

func StructToString(s interface{}) (string, error) {
	if s == nil {
		return "", errors.New("nil struct")
	}
	b, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("failed to marshal struct: %w", err)
	}
	return string(b), nil
}
