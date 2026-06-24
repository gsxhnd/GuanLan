package task

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gsxhnd/guanlan/internal/data"
)

// PythonSyncConfig Python 日频同步进程配置。
type PythonSyncConfig struct {
	PythonBin string
	RepoRoot  string
	DBPath    string
}

// PythonDataSyncExecutor 调用 Python daily_data 服务入库。
type PythonDataSyncExecutor struct {
	Store  *data.Store
	Config PythonSyncConfig
}

func (e *PythonDataSyncExecutor) Run(ctx context.Context, task data.DataSyncTask) error {
	if task.TaskType != data.TaskTypeDataSync {
		return fmt.Errorf("unsupported task type: %s", task.TaskType)
	}

	pythonBin := e.Config.PythonBin
	if pythonBin == "" {
		pythonBin = "uv"
	}
	repoRoot := e.Config.RepoRoot
	if repoRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		repoRoot = wd
	}
	dbPath := e.Config.DBPath
	if dbPath == "" {
		dbPath = data.DefaultDBPath
	}
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(repoRoot, dbPath)
	}

	var cmd *exec.Cmd
	if pythonBin == "uv" {
		cmd = exec.CommandContext(
			ctx,
			"uv", "run", "python", "-m", "services.daily_data",
			"sync", "--db", dbPath, "--stock", task.TargetObject,
		)
	} else {
		cmd = exec.CommandContext(
			ctx,
			pythonBin, "-m", "services.daily_data",
			"sync", "--db", dbPath, "--stock", task.TargetObject,
		)
	}
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "PYTHONPATH="+repoRoot)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		_ = e.Store.UpdateTaskStatus(ctx, task.TaskID, data.TaskStatusFailed, &msg, nil)
		_ = e.Store.UpsertStockDataStatus(ctx, data.StockDataStatus{
			StockCode:  task.TargetObject,
			StockName:  task.TargetObject,
			Market:     inferMarket(task.TargetObject),
			SyncStatus: data.StockStatusMissing,
		})
		return fmt.Errorf("python sync: %s", msg)
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		version = fmt.Sprintf("v%s", task.CreatedAt.Format("20060102"))
	}
	v := version
	if err := e.Store.UpdateTaskStatus(ctx, task.TaskID, data.TaskStatusSuccess, nil, &v); err != nil {
		return err
	}
	return nil
}

func inferMarket(code string) data.Market {
	entry, err := data.ParsePoolSymbol(code, "")
	if err != nil {
		return data.MarketUS
	}
	return entry.Market
}

// RunDailySync 对 stock_pool 执行增量日频同步（近 lookbackDays 天）。
func RunDailySync(ctx context.Context, cfg PythonSyncConfig, lookbackDays int) error {
	if lookbackDays <= 0 {
		lookbackDays = 7
	}
	return runPythonModule(ctx, cfg, "daily-sync", "--lookback-days", fmt.Sprintf("%d", lookbackDays))
}

// InitTrainingData 兼容旧初始化流程。
func InitTrainingData(ctx context.Context, cfg PythonSyncConfig, indexCodes ...string) error {
	args := []string{"init-training"}
	for _, code := range indexCodes {
		args = append(args, "--index-code", code)
	}
	return runPythonModule(ctx, cfg, args...)
}

func runPythonModule(ctx context.Context, cfg PythonSyncConfig, args ...string) error {
	pythonBin := cfg.PythonBin
	if pythonBin == "" {
		pythonBin = "uv"
	}
	repoRoot := cfg.RepoRoot
	if repoRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		repoRoot = wd
	}
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = data.DefaultDBPath
	}
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(repoRoot, dbPath)
	}

	baseArgs := []string{"run", "python", "-m", "services.daily_data", "--db", dbPath}
	if pythonBin != "uv" {
		baseArgs = []string{"-m", "services.daily_data", "--db", dbPath}
	}
	baseArgs = append(baseArgs, args...)

	var cmd *exec.Cmd
	if pythonBin == "uv" {
		cmd = exec.CommandContext(ctx, "uv", baseArgs...)
	} else {
		cmd = exec.CommandContext(ctx, pythonBin, baseArgs...)
	}
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "PYTHONPATH="+repoRoot)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s: %s", args[0], msg)
	}
	return nil
}
