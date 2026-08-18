package orchestrator

import (
	"context"
	"fmt"

	"github.com/gsxhnd/guanlan/internal/data"
)

// Dispatcher routes claimed tasks by type.
type Dispatcher struct {
	Sync    *SyncExecutor
	Predict *PredictExecutor
}

func (d *Dispatcher) Run(ctx context.Context, task data.DataSyncTask) error {
	switch task.TaskType {
	case data.TaskTypeDataSync:
		if d.Sync == nil {
			return fmt.Errorf("sync executor not configured")
		}
		return d.Sync.Run(ctx, task)
	case data.TaskTypeAnalysis:
		if d.Predict == nil {
			return fmt.Errorf("predict executor not configured")
		}
		return d.Predict.Run(ctx, task)
	default:
		return fmt.Errorf("unsupported task type: %s", task.TaskType)
	}
}
