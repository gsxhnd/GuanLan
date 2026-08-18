package main

import (
	"flag"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
)

type Config struct {
	HTTPAddr         string
	DBPath           string
	CrawlerAddr      string
	PredictionAddr   string
	SyncLookbackDays int
	TaskPollInterval time.Duration
	DailySyncCron    string
}

func NewConfig() Config {
	httpAddr := flag.String("http", ":8080", "HTTP listen address")
	dbPath := flag.String("db", data.DefaultDBPath, "DuckDB file path")
	crawlerAddr := flag.String("crawler", "localhost:50061", "crawler gRPC address")
	predictionAddr := flag.String("prediction", "localhost:50062", "prediction gRPC address")
	lookback := flag.Int("sync-lookback-days", 7, "daily sync lookback window in days")
	taskPoll := flag.Duration("task-poll", 2*time.Second, "task scheduler poll interval")
	dailyCron := flag.String("daily-sync-cron", "0 18 * * *", "robfig cron spec for daily sync enqueue")
	flag.Parse()

	return Config{
		HTTPAddr:         *httpAddr,
		DBPath:           *dbPath,
		CrawlerAddr:      *crawlerAddr,
		PredictionAddr:   *predictionAddr,
		SyncLookbackDays: *lookback,
		TaskPollInterval: *taskPoll,
		DailySyncCron:    *dailyCron,
	}
}
