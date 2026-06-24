package main

import (
	"flag"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
)

type Config struct {
	GRPCAddr              string
	DBPath                string
	RepoRoot              string
	PythonBin             string
	InitTraining          bool
	TaskPollInterval      time.Duration
	ScheduledSyncInterval time.Duration
}

func NewConfig() Config {
	grpcAddr := flag.String("addr", ":50051", "gRPC listen address")
	dbPath := flag.String("db", data.DefaultDBPath, "DuckDB file path")
	repoRoot := flag.String("repo-root", ".", "repository root for Python services")
	pythonBin := flag.String("python", "uv", "Python runner (uv or python3)")
	initTraining := flag.Bool("init-training", false, "initialize training index data on startup")
	taskPoll := flag.Duration("task-poll", 2*time.Second, "task scheduler poll interval")
	scheduledSync := flag.Duration("scheduled-sync", time.Hour, "scheduled data sync interval")
	flag.Parse()

	return Config{
		GRPCAddr:              *grpcAddr,
		DBPath:                *dbPath,
		RepoRoot:              *repoRoot,
		PythonBin:             *pythonBin,
		InitTraining:          *initTraining,
		TaskPollInterval:      *taskPoll,
		ScheduledSyncInterval: *scheduledSync,
	}
}
