package models

import (
	"time"

	"github.com/google/uuid"
)

type ClickEvent struct {
	ID           uuid.UUID `json:"id"`
	URLID        uuid.UUID `json:"url_id"`
	ShortCode    string    `json:"short_code"`
	ClickedAt    time.Time `json:"clicked_at"`
	Country      *string   `json:"country,omitempty"`
	City         *string   `json:"city,omitempty"`
	DeviceType   string    `json:"device_type"`
	Browser      string    `json:"browser"`
	OS           string    `json:"os"`
	Referrer     *string   `json:"referrer,omitempty"`
	ReferrerHost *string   `json:"referrer_host,omitempty"`
	IPHash       string    `json:"ip_hash"`
}

type AnalyticsQuery struct {
	From     time.Time
	To       time.Time
	Interval string
}

type TimeSeriesPoint struct {
	Timestamp string `json:"timestamp"`
	Clicks    int64  `json:"clicks"`
}

type DistributionItem struct {
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type AnalyticsSummary struct {
	ShortCode        string             `json:"short_code"`
	TotalClicks      int64              `json:"total_clicks"`
	UniqueVisitors   int64              `json:"unique_visitors"`
	From             time.Time          `json:"from"`
	To               time.Time          `json:"to"`
	ClicksOverTime   []TimeSeriesPoint  `json:"clicks_over_time"`
	Countries        []DistributionItem `json:"countries"`
	Devices          []DistributionItem `json:"devices"`
	Browsers         []DistributionItem `json:"browsers"`
	OperatingSystems []DistributionItem `json:"operating_systems"`
	Referrers        []DistributionItem `json:"referrers"`
}
