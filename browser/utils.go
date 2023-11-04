package browser

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func IsValidURL(URL string) (bool, error) {
	if URL == "" {
		return false, errors.New("url cannot be empty")
	} else if _, err := url.ParseRequestURI(URL); err != nil {
		return false, fmt.Errorf("error parsing url: %w", err)
	}
	return true, nil
}

// Note: This only works for HTTP and HTTPS URLs.
func GetCanonicalURL(URL string) (string, error) {
	if !strings.HasPrefix(URL, "http://") && !strings.HasPrefix(URL, "https://") {
		URL = "https://" + URL
	}
	resp, err := http.Head(URL)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc, err := resp.Location()
		if err != nil {
			return "", err
		}
		return loc.String(), nil
	}
	if !strings.Contains(URL, "://www.") {
		wwwVersion := strings.Replace(URL, "://", "://www.", 1)
		resp, err := http.Head(wwwVersion)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 400 {
			resp.Body.Close()
			return wwwVersion, nil
		}
	}
	return URL, nil
}
