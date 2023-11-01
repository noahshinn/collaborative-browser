package utils

import (
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

func WriteFile(filepath string, text string) error {
	return os.WriteFile(filepath, []byte(text), 0644)
}
