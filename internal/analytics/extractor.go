package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/models"
)

type Extractor interface {
	Extract(r *http.Request, urlID uuid.UUID, shortCode string) models.ClickEvent
}

type DefaultExtractor struct{}

func NewExtractor() *DefaultExtractor {
	return &DefaultExtractor{}
}

func (e *DefaultExtractor) Extract(r *http.Request, urlID uuid.UUID, shortCode string) models.ClickEvent {
	ip := extractClientIP(r)
	ipHash := hashIP(ip)

	userAgent := r.UserAgent()
	deviceType := detectDeviceType(userAgent)
	browser := detectBrowser(userAgent)
	os := detectOS(userAgent)

	rawReferrer := r.Referer()
	var referrer *string
	var referrerHost *string
	if rawReferrer != "" {
		referrer = &rawReferrer
		host := extractReferrerHost(rawReferrer)
		if host != "" {
			referrerHost = &host
		}
	}

	country := extractCountry(r)
	city := extractCity(r)

	return models.ClickEvent{
		ID:           uuid.New(),
		URLID:        urlID,
		ShortCode:    shortCode,
		ClickedAt:    time.Now().UTC(),
		Country:      country,
		City:         city,
		DeviceType:   deviceType,
		Browser:      browser,
		OS:           os,
		Referrer:     referrer,
		ReferrerHost: referrerHost,
		IPHash:       ipHash,
	}
}

func extractClientIP(r *http.Request) string {
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		return cfIP
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xRealIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); xRealIP != "" {
		return xRealIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func hashIP(ip string) string {
	if ip == "" {
		ip = "127.0.0.1"
	}
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:])
}

func extractCountry(r *http.Request) *string {
	headers := []string{"CF-IPCountry", "X-Country-Code", "X-GeoIP-Country"}
	for _, h := range headers {
		if val := strings.TrimSpace(r.Header.Get(h)); val != "" && val != "XX" && val != "T1" {
			upper := strings.ToUpper(val)
			return &upper
		}
	}
	return nil
}

func extractCity(r *http.Request) *string {
	headers := []string{"CF-IPCity", "X-GeoIP-City", "X-City"}
	for _, h := range headers {
		if val := strings.TrimSpace(r.Header.Get(h)); val != "" {
			return &val
		}
	}
	return nil
}

func extractReferrerHost(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "direct"
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return "direct"
	}
	return host
}

func detectDeviceType(ua string) string {
	lower := strings.ToLower(ua)
	if lower == "" {
		return "unknown"
	}
	if strings.Contains(lower, "bot") || strings.Contains(lower, "crawler") || strings.Contains(lower, "spider") || strings.Contains(lower, "curl") || strings.Contains(lower, "k6") {
		return "bot"
	}
	if strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet") || strings.Contains(lower, "kindle") {
		return "tablet"
	}
	if strings.Contains(lower, "mobile") || strings.Contains(lower, "iphone") || strings.Contains(lower, "android") {
		return "mobile"
	}
	return "desktop"
}

func detectBrowser(ua string) string {
	lower := strings.ToLower(ua)
	if lower == "" {
		return "Other"
	}
	if strings.Contains(lower, "edg/") || strings.Contains(lower, "edge/") {
		return "Edge"
	}
	if strings.Contains(lower, "opr/") || strings.Contains(lower, "opera") {
		return "Opera"
	}
	if strings.Contains(lower, "chrome/") || strings.Contains(lower, "crios/") {
		return "Chrome"
	}
	if strings.Contains(lower, "firefox/") || strings.Contains(lower, "fxios/") {
		return "Firefox"
	}
	if strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome") {
		return "Safari"
	}
	return "Other"
}

func detectOS(ua string) string {
	lower := strings.ToLower(ua)
	if lower == "" {
		return "Other"
	}
	if strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ios") {
		return "iOS"
	}
	if strings.Contains(lower, "android") {
		return "Android"
	}
	if strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os x") {
		return "macOS"
	}
	if strings.Contains(lower, "windows") {
		return "Windows"
	}
	if strings.Contains(lower, "linux") {
		return "Linux"
	}
	return "Other"
}
