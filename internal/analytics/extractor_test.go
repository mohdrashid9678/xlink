package analytics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExtractorParsesHeadersAndClientInfo(t *testing.T) {
	extractor := NewExtractor()
	urlID := uuid.New()
	shortCode := "test1234"

	req := httptest.NewRequest(http.MethodGet, "/"+shortCode, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1")
	req.Header.Set("Referer", "https://t.co/abc123xyz")
	req.Header.Set("CF-Connecting-IP", "203.0.113.195")
	req.Header.Set("CF-IPCountry", "IN")
	req.Header.Set("CF-IPCity", "Mumbai")

	event := extractor.Extract(req, urlID, shortCode)

	assert.Equal(t, urlID, event.URLID)
	assert.Equal(t, shortCode, event.ShortCode)
	assert.Equal(t, "mobile", event.DeviceType)
	assert.Equal(t, "Safari", event.Browser)
	assert.Equal(t, "iOS", event.OS)
	assert.NotNil(t, event.Country)
	assert.Equal(t, "IN", *event.Country)
	assert.NotNil(t, event.City)
	assert.Equal(t, "Mumbai", *event.City)
	assert.NotNil(t, event.ReferrerHost)
	assert.Equal(t, "t.co", *event.ReferrerHost)
	assert.NotEmpty(t, event.IPHash)
}

func TestExtractorDetectsDifferentBrowsersAndOS(t *testing.T) {
	extractor := NewExtractor()
	urlID := uuid.New()

	tests := []struct {
		name       string
		ua         string
		wantDevice string
		wantBrowser string
		wantOS     string
	}{
		{
			name:        "Chrome on Windows",
			ua:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
			wantDevice:  "desktop",
			wantBrowser: "Chrome",
			wantOS:      "Windows",
		},
		{
			name:        "Firefox on macOS",
			ua:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 14.3; rv:123.0) Gecko/20100101 Firefox/123.0",
			wantDevice:  "desktop",
			wantBrowser: "Firefox",
			wantOS:      "macOS",
		},
		{
			name:        "Edge on Windows",
			ua:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.2365.66",
			wantDevice:  "desktop",
			wantBrowser: "Edge",
			wantOS:      "Windows",
		},
		{
			name:        "Android Mobile Chrome",
			ua:          "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.64 Mobile Safari/537.36",
			wantDevice:  "mobile",
			wantBrowser: "Chrome",
			wantOS:      "Android",
		},
		{
			name:        "iPad Tablet Safari",
			ua:          "Mozilla/5.0 (iPad; CPU OS 17_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Mobile/15E148 Safari/604.1",
			wantDevice:  "tablet",
			wantBrowser: "Safari",
			wantOS:      "iOS",
		},
		{
			name:        "Bot / Crawler",
			ua:          "Googlebot/2.1 (+http://www.google.com/bot.html)",
			wantDevice:  "bot",
			wantBrowser: "Other",
			wantOS:      "Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/abc", nil)
			req.Header.Set("User-Agent", tt.ua)
			event := extractor.Extract(req, urlID, "abc")

			assert.Equal(t, tt.wantDevice, event.DeviceType)
			assert.Equal(t, tt.wantBrowser, event.Browser)
			assert.Equal(t, tt.wantOS, event.OS)
		})
	}
}
