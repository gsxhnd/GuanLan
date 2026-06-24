package main

import (
	"flag"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
)

type Config struct {
	GRPCAddr              string
	DBPath                string
	TaskPollInterval      time.Duration
	ScheduledSyncInterval time.Duration
}

func NewConfig() Config {
	grpcAddr := flag.String("addr", ":50051", "gRPC listen address")
	dbPath := flag.String("db", data.DefaultDBPath, "DuckDB file path")
	taskPoll := flag.Duration("task-poll", 2*time.Second, "task scheduler poll interval")
	scheduledSync := flag.Duration("scheduled-sync", time.Hour, "scheduled data sync interval")
	flag.Parse()

	return Config{
		GRPCAddr:              *grpcAddr,
		DBPath:                *dbPath,
		TaskPollInterval:      *taskPoll,
		ScheduledSyncInterval: *scheduledSync,
	}
}
