package io

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadFile(filepath string) (string, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	return string(bytes), nil
}

func WriteBytesToFile(filepath string, bytes []byte) error {
	return os.WriteFile(filepath, bytes, 0644)
}

func WriteStringToFile(filepath string, str string) error {
	return WriteBytesToFile(filepath, []byte(str))
}

func WriteStructToFile(filepath string, object interface{}) error {
	if bytes, err := json.Marshal(object); err != nil {
		return fmt.Errorf("error marshalling object: %w", err)
	} else {
		return WriteBytesToFile(filepath, bytes)
	}
}
