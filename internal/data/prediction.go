package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Prediction is a persisted inference row.
type Prediction struct {
	PredictionID string
	StockCode    string
	TradeDate    time.Time
	Score        float64
	ModelVersion string
	CreatedAt    time.Time
}

// ModelVersion is a registered model artifact.
type ModelVersion struct {
	ModelVersion  string
	Description   string
	CreatedAt     time.Time
	ArtifactPath  string
}

// InsertPredictions upserts scores by (stock_code, trade_date, model_version).
func (s *Store) InsertPredictions(ctx context.Context, rows []Prediction) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	for i := range rows {
		row := &rows[i]
		if row.StockCode == "" {
			return fmt.Errorf("stock_code is required")
		}
		row.StockCode = strings.ToUpper(strings.TrimSpace(row.StockCode))
		if row.PredictionID == "" {
			row.PredictionID = uuid.NewString()
		}
		if row.CreatedAt.IsZero() {
			row.CreatedAt = now
		}
		if row.ModelVersion == "" {
			row.ModelVersion = "baseline"
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO predictions (
				prediction_id, stock_code, trade_date, score, model_version, created_at
			) VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (stock_code, trade_date, model_version) DO UPDATE SET
				score = excluded.score,
				created_at = excluded.created_at
		`, row.PredictionID, row.StockCode, row.TradeDate, row.Score, row.ModelVersion, row.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert prediction: %w", err)
		}
	}
	return tx.Commit()
}

// ListPredictions filters stored predictions.
func (s *Store) ListPredictions(ctx context.Context, stockCode, tradeDate, modelVersion string, limit int) ([]Prediction, error) {
	if limit <= 0 {
		limit = 100
	}
	var (
		args  []any
		where []string
	)
	if stockCode != "" {
		where = append(where, "stock_code = ?")
		args = append(args, strings.ToUpper(strings.TrimSpace(stockCode)))
	}
	if tradeDate != "" {
		where = append(where, "trade_date = ?")
		args = append(args, tradeDate)
	}
	if modelVersion != "" {
		where = append(where, "model_version = ?")
		args = append(args, modelVersion)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT prediction_id, stock_code, trade_date, score, model_version, created_at
		FROM predictions
		%s
		ORDER BY trade_date DESC, stock_code ASC
		LIMIT ?
	`, whereSQL), args...)
	if err != nil {
		return nil, fmt.Errorf("list predictions: %w", err)
	}
	defer rows.Close()

	var out []Prediction
	for rows.Next() {
		var p Prediction
		if err := rows.Scan(&p.PredictionID, &p.StockCode, &p.TradeDate, &p.Score, &p.ModelVersion, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertModelVersion registers a model artifact path.
func (s *Store) UpsertModelVersion(ctx context.Context, mv ModelVersion) error {
	if mv.ModelVersion == "" {
		return fmt.Errorf("model_version is required")
	}
	if mv.CreatedAt.IsZero() {
		mv.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO model_versions (model_version, description, created_at, artifact_path)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (model_version) DO UPDATE SET
			description = excluded.description,
			artifact_path = excluded.artifact_path
	`, mv.ModelVersion, nullString(mv.Description), mv.CreatedAt, nullString(mv.ArtifactPath))
	if err != nil {
		return fmt.Errorf("upsert model version: %w", err)
	}
	return nil
}

// LatestModelVersion returns the most recently created model version, if any.
func (s *Store) LatestModelVersion(ctx context.Context) (ModelVersion, bool, error) {
	var mv ModelVersion
	var desc, path sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT model_version, description, created_at, artifact_path
		FROM model_versions
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&mv.ModelVersion, &desc, &mv.CreatedAt, &path)
	if err == sql.ErrNoRows {
		return ModelVersion{}, false, nil
	}
	if err != nil {
		return ModelVersion{}, false, fmt.Errorf("latest model version: %w", err)
	}
	if desc.Valid {
		mv.Description = desc.String
	}
	if path.Valid {
		mv.ArtifactPath = path.String
	}
	return mv, true, nil
}
