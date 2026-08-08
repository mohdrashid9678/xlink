package service

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const maxShortCodeLength = 64

var shortCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func validateLongURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("long_url must be an absolute HTTP(S) URL: %w", ErrValidation)
	}
	return value, nil
}

func validateShortCode(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !shortCodePattern.MatchString(value) {
		return "", fmt.Errorf("short code must contain 1-%d letters, numbers, underscores, or hyphens: %w", maxShortCodeLength, ErrValidation)
	}
	return value, nil
}

func validateExpiry(value *time.Time) error {
	if value != nil && !value.After(time.Now().UTC()) {
		return fmt.Errorf("expires_at must be in the future: %w", ErrValidation)
	}
	return nil
}
