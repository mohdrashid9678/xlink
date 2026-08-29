package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohdrashid9678/xlink/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type AnalyticsRepository interface {
	BatchInsertClicks(ctx context.Context, clicks []models.ClickEvent) error
	GetAnalyticsSummary(ctx context.Context, urlID uuid.UUID, shortCode string, query models.AnalyticsQuery) (*models.AnalyticsSummary, error)
}

type PostgresAnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAnalyticsRepository(pool *pgxpool.Pool) *PostgresAnalyticsRepository {
	return &PostgresAnalyticsRepository{pool: pool}
}

func (r *PostgresAnalyticsRepository) BatchInsertClicks(ctx context.Context, clicks []models.ClickEvent) error {
	if len(clicks) == 0 {
		return nil
	}

	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.BatchInsertClicks",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	batch := &pgx.Batch{}

	// 1. Queue multi-row inserts for clicks
	valueStrings := make([]string, 0, len(clicks))
	valueArgs := make([]any, 0, len(clicks)*12)

	urlClickCounts := make(map[uuid.UUID]int64)

	for i, c := range clicks {
		offset := i * 12
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6,
			offset+7, offset+8, offset+9, offset+10, offset+11, offset+12,
		))
		valueArgs = append(valueArgs,
			c.ID, c.URLID, c.ShortCode, c.ClickedAt,
			c.Country, c.City, c.DeviceType, c.Browser, c.OS,
			c.Referrer, c.ReferrerHost, c.IPHash,
		)
		urlClickCounts[c.URLID]++
	}

	insertQuery := fmt.Sprintf(
		`INSERT INTO clicks (id, url_id, short_code, clicked_at, country, city, device_type, browser, os, referrer, referrer_host, ip_hash) VALUES %s`,
		strings.Join(valueStrings, ", "),
	)
	batch.Queue(insertQuery, valueArgs...)

	// 2. Queue batch counter updates on urls table
	for urlID, count := range urlClickCounts {
		batch.Queue(
			`UPDATE urls SET click_count = click_count + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
			count, urlID,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return mapDatabaseError(err)
		}
	}

	return nil
}

func (r *PostgresAnalyticsRepository) GetAnalyticsSummary(ctx context.Context, urlID uuid.UUID, shortCode string, query models.AnalyticsQuery) (*models.AnalyticsSummary, error) {
	ctx, span := otel.Tracer("xlink-db").Start(ctx, "db.GetAnalyticsSummary",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	from := query.From
	to := query.To
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if from.IsZero() {
		from = to.AddDate(0, 0, -30) // default last 30 days
	}

	interval := query.Interval
	if interval != "hour" && interval != "day" {
		interval = "day"
	}

	summary := &AnalyticsSummaryData{
		ShortCode:        shortCode,
		From:             from,
		To:               to,
		ClicksOverTime:   make([]models.TimeSeriesPoint, 0),
		Countries:        make([]models.DistributionItem, 0),
		Devices:          make([]models.DistributionItem, 0),
		Browsers:         make([]models.DistributionItem, 0),
		OperatingSystems: make([]models.DistributionItem, 0),
		Referrers:        make([]models.DistributionItem, 0),
	}

	// 1. Total Clicks and Unique Visitors
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT ip_hash)
		FROM clicks
		WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at <= $3
	`, urlID, from, to).Scan(&summary.TotalClicks, &summary.UniqueVisitors)
	if err != nil {
		return nil, mapDatabaseError(err)
	}

	if summary.TotalClicks == 0 {
		return summary.ToModel(), nil
	}

	// 2. Clicks Over Time
	truncFormat := "day"
	if interval == "hour" {
		truncFormat = "hour"
	}
	timeRows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT date_trunc('%s', clicked_at) as bucket, COUNT(*)
		FROM clicks
		WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at <= $3
		GROUP BY bucket
		ORDER BY bucket ASC
	`, truncFormat), urlID, from, to)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	defer timeRows.Close()

	for timeRows.Next() {
		var bucket time.Time
		var count int64
		if err := timeRows.Scan(&bucket, &count); err != nil {
			return nil, mapDatabaseError(err)
		}
		summary.ClicksOverTime = append(summary.ClicksOverTime, models.TimeSeriesPoint{
			Timestamp: bucket.UTC().Format(time.RFC3339),
			Clicks:    count,
		})
	}

	// Helper for distribution breakdown
	queryDistribution := func(column string, limit int) ([]models.DistributionItem, error) {
		q := fmt.Sprintf(`
			SELECT COALESCE(%s, 'Unknown') as label, COUNT(*) as cnt
			FROM clicks
			WHERE url_id = $1 AND clicked_at >= $2 AND clicked_at <= $3
			GROUP BY label
			ORDER BY cnt DESC
			LIMIT %d
		`, column, limit)

		rows, err := r.pool.Query(ctx, q, urlID, from, to)
		if err != nil {
			return nil, mapDatabaseError(err)
		}
		defer rows.Close()

		items := make([]models.DistributionItem, 0)
		for rows.Next() {
			var label string
			var count int64
			if err := rows.Scan(&label, &count); err != nil {
				return nil, mapDatabaseError(err)
			}
			percentage := float64(0)
			if summary.TotalClicks > 0 {
				percentage = float64(count) / float64(summary.TotalClicks) * 100.0
			}
			items = append(items, models.DistributionItem{
				Name:       label,
				Count:      count,
				Percentage: percentage,
			})
		}
		return items, nil
	}

	// 3. Countries
	if summary.Countries, err = queryDistribution("country", 10); err != nil {
		return nil, err
	}
	// 4. Devices
	if summary.Devices, err = queryDistribution("device_type", 10); err != nil {
		return nil, err
	}
	// 5. Browsers
	if summary.Browsers, err = queryDistribution("browser", 10); err != nil {
		return nil, err
	}
	// 6. OS
	if summary.OperatingSystems, err = queryDistribution("os", 10); err != nil {
		return nil, err
	}
	// 7. Referrers
	if summary.Referrers, err = queryDistribution("referrer_host", 10); err != nil {
		return nil, err
	}

	return summary.ToModel(), nil
}

type AnalyticsSummaryData struct {
	ShortCode        string
	TotalClicks      int64
	UniqueVisitors   int64
	From             time.Time
	To               time.Time
	ClicksOverTime   []models.TimeSeriesPoint
	Countries        []models.DistributionItem
	Devices          []models.DistributionItem
	Browsers         []models.DistributionItem
	OperatingSystems []models.DistributionItem
	Referrers        []models.DistributionItem
}

func (d *AnalyticsSummaryData) ToModel() *models.AnalyticsSummary {
	return &models.AnalyticsSummary{
		ShortCode:        d.ShortCode,
		TotalClicks:      d.TotalClicks,
		UniqueVisitors:   d.UniqueVisitors,
		From:             d.From,
		To:               d.To,
		ClicksOverTime:   d.ClicksOverTime,
		Countries:        d.Countries,
		Devices:          d.Devices,
		Browsers:         d.Browsers,
		OperatingSystems: d.OperatingSystems,
		Referrers:        d.Referrers,
	}
}
