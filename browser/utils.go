package browser

import (
	"errors"
	"fmt"
	"net/url"
)

func IsValidURL(u string) (bool, error) {
	if u == "" {
		return false, errors.New("url cannot be empty")
	} else if _, err := url.ParseRequestURI(u); err != nil {
		return false, fmt.Errorf("error parsing url: %w", err)
	}
	return true, nil
}
