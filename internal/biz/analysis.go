package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
)

const AnalysisTargetWatchlist = "*"

// EncodeAnalysisTarget stores trade_date|model_version|codes for an analysis task.
func EncodeAnalysisTarget(tradeDate, modelVersion string, stockCodes []string) string {
	if tradeDate == "" {
		tradeDate = time.Now().UTC().Format("2006-01-02")
	}
	if modelVersion == "" {
		modelVersion = "latest"
	}
	codes := AnalysisTargetWatchlist
	if len(stockCodes) > 0 {
		out := make([]string, 0, len(stockCodes))
		for _, c := range stockCodes {
			c = strings.ToUpper(strings.TrimSpace(c))
			if c != "" {
				out = append(out, c)
			}
		}
		if len(out) > 0 {
			codes = strings.Join(out, ",")
		}
	}
	return tradeDate + "|" + modelVersion + "|" + codes
}

// DecodeAnalysisTarget parses EncodeAnalysisTarget output.
func DecodeAnalysisTarget(target string) (tradeDate, modelVersion string, codes []string) {
	parts := strings.SplitN(target, "|", 3)
	if len(parts) != 3 {
		return time.Now().UTC().Format("2006-01-02"), "latest", []string{AnalysisTargetWatchlist}
	}
	return parts[0], parts[1], strings.Split(parts[2], ",")
}

// CreateAnalysisTask enqueues a pending analysis (batch or watchlist).
func (t *Task) CreateAnalysisTask(ctx context.Context, tradeDate, modelVersion string, stockCodes []string, trigger data.TriggerMethod) (data.DataSyncTask, error) {
	if trigger == "" {
		trigger = data.TriggerManual
	}
	target := EncodeAnalysisTarget(tradeDate, modelVersion, stockCodes)
	task, err := t.Store.CreateTask(ctx, data.TaskTypeAnalysis, target, trigger, 0)
	if err != nil {
		return data.DataSyncTask{}, fmt.Errorf("create analysis task: %w", err)
	}
	return task, nil
}
