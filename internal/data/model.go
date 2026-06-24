package data

import "time"

// Market 交易市场。
type Market string

const (
	MarketA  Market = "A"
	MarketUS Market = "US"
)

// StockSyncStatus 个股数据同步状态，与前端 data-page 状态标签一致。
type StockSyncStatus string

const (
	StockStatusReady   StockSyncStatus = "ready"
	StockStatusSyncing StockSyncStatus = "syncing"
	StockStatusMissing StockSyncStatus = "missing"
)

// TaskType 任务类型。
type TaskType string

const (
	TaskTypeDataSync TaskType = "data_sync"
)

// TriggerMethod 任务触发方式。
type TriggerMethod string

const (
	TriggerScheduled TriggerMethod = "scheduled"
	TriggerManual    TriggerMethod = "manual"
	TriggerRetry     TriggerMethod = "retry"
)

// TaskStatus 统一任务状态，见 docs/dev/03-domain-model.md §1。
type TaskStatus string

const (
	TaskStatusPending        TaskStatus = "pending"
	TaskStatusRunning        TaskStatus = "running"
	TaskStatusSuccess        TaskStatus = "success"
	TaskStatusFailed         TaskStatus = "failed"
	TaskStatusPartialSuccess TaskStatus = "partial_success"
	TaskStatusCancelled      TaskStatus = "cancelled"
)

// IndexDataset 预置训练指数数据集状态（§4.1）。
type IndexDataset struct {
	IndexCode        string     `json:"index_code"`
	Market           Market     `json:"market"`
	IndexName        string     `json:"index_name"`
	DataCompleteness float64    `json:"data_completeness"`
	LastSyncTime     *time.Time `json:"last_sync_time,omitempty"`
	SyncStatus       string     `json:"sync_status"`
}

// IndexConstituent 指数成分股快照（§4.2）。
type IndexConstituent struct {
	IndexCode string     `json:"index_code"`
	StockCode string     `json:"stock_code"`
	SnapDate  time.Time  `json:"snap_date"`
	Weight    *float64   `json:"weight,omitempty"`
	IsActive  bool       `json:"is_active"`
}

// DailyBar 日频行情（§4.3）。
type DailyBar struct {
	StockCode   string    `json:"stock_code"`
	Market      Market    `json:"market"`
	TradeDate   time.Time `json:"trade_date"`
	Open        float64   `json:"open"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Close       float64   `json:"close"`
	Volume      int64     `json:"volume"`
	Amount      *float64  `json:"amount,omitempty"`
	AdjFactor   *float64  `json:"adj_factor,omitempty"`
	Source      string    `json:"source"`
	DataVersion string    `json:"data_version"`
}

// DateRange 缺失数据区间。
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// StockDataStatus 个股数据状态（§4.4）。
type StockDataStatus struct {
	StockCode         string          `json:"stock_code"`
	StockName         string          `json:"stock_name"`
	Market            Market          `json:"market"`
	TrainingIndexCode *string         `json:"training_index_code,omitempty"`
	DataStartDate     *time.Time      `json:"data_start_date,omitempty"`
	DataEndDate       *time.Time      `json:"data_end_date,omitempty"`
	Completeness      float64         `json:"completeness"`
	MissingRanges     []DateRange     `json:"missing_ranges,omitempty"`
	LastUpdate        *time.Time      `json:"last_update,omitempty"`
	SyncStatus        StockSyncStatus `json:"sync_status"`
}

// StockListItem 数据页股票列表行：状态 + 最近一根日 K 摘要。
type StockListItem struct {
	StockDataStatus
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Change float64 `json:"change"`
}

// DataSyncTask 数据同步任务（§2，task_type=data_sync）。
type DataSyncTask struct {
	TaskID        string        `json:"task_id"`
	TaskType      TaskType      `json:"task_type"`
	TargetObject  string        `json:"target_object"`
	TriggerMethod TriggerMethod `json:"trigger_method"`
	Status        TaskStatus    `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	StartedAt     *time.Time    `json:"started_at,omitempty"`
	EndedAt       *time.Time    `json:"ended_at,omitempty"`
	RetryCount    int           `json:"retry_count"`
	FailureReason *string       `json:"failure_reason,omitempty"`
	LogRef        *string       `json:"log_ref,omitempty"`
}

// ListStocksFilter 对应 GET /api/data/stocks 查询参数。
type ListStocksFilter struct {
	Market *Market
	Status *StockSyncStatus
	Search string
	Sort   string // code | name | change | volume
}

// ListDailyBarsParams 对应 GET /api/data/stocks/{stock_code}/daily-bars。
type ListDailyBarsParams struct {
	StockCode string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
}
