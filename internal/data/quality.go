package data

import (
	"context"
	"fmt"
)

// StockQualitySummary 个股质量摘要。
type StockQualitySummary struct {
	StockCode     string `json:"stock_code"`
	IssueCount    int    `json:"issue_count"`
	CriticalCount int    `json:"critical_count"`
	QualityStatus string `json:"quality_status"` // ok | warn | critical
}

// GetStockQualitySummary 返回个股质量状态摘要。
func (s *Store) GetStockQualitySummary(ctx context.Context, stockCode string) (StockQualitySummary, error) {
	var critical, total int
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)::INTEGER,
			COUNT(*) FILTER (WHERE severity = 'critical')::INTEGER
		FROM data_quality_issues
		WHERE stock_code = ?
	`, stockCode).Scan(&total, &critical)
	if err != nil {
		return StockQualitySummary{}, fmt.Errorf("stock quality summary: %w", err)
	}

	status := "ok"
	if critical > 0 {
		status = "critical"
	} else if total > 0 {
		status = "warn"
	}
	return StockQualitySummary{
		StockCode:     stockCode,
		IssueCount:    total,
		CriticalCount: critical,
		QualityStatus: status,
	}, nil
}
