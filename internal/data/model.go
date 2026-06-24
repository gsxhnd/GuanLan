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
	TaskTypeAnalysis TaskType = "analysis"
	TaskTypeTraining TaskType = "training"
	TaskTypeBacktest TaskType = "backtest"
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

// DataSyncTask 统一任务记录（§2）。
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
	DataVersion   *string       `json:"data_version,omitempty"`
}

// StockPoolSource 股票池条目来源。
type StockPoolSource string

const (
	PoolSourceCSVImport StockPoolSource = "csv_import"
	PoolSourceAPIManual StockPoolSource = "api_manual"
)

// StockPoolEntry DuckDB 数据底座股票池（日频拉取范围）。
type StockPoolEntry struct {
	YfinanceSymbol string          `json:"yfinance_symbol"`
	OriginalCode   string          `json:"original_code"`
	Market         Market          `json:"market"`
	StockName      string          `json:"stock_name"`
	Exchange       string          `json:"exchange,omitempty"`
	Currency       string          `json:"currency"`
	Source         StockPoolSource `json:"source"`
	IsActive       bool            `json:"is_active"`
	SyncDaily      bool            `json:"sync_daily"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// WatchlistItem 用户关注股票池条目（§4.5）。
type WatchlistItem struct {
	StockCode    string     `json:"stock_code"`
	Market       Market     `json:"market"`
	Tags         []string   `json:"tags,omitempty"`
	Priority     int        `json:"priority"`
	Notes        string     `json:"notes,omitempty"`
	IsActive     bool       `json:"is_active"`
	AddedAt      time.Time  `json:"added_at"`
	RemovedAt    *time.Time `json:"removed_at,omitempty"`
	Source       string     `json:"source"`
	LastAction   string     `json:"last_action,omitempty"`
	LastActionAt *time.Time `json:"last_action_at,omitempty"`
	SyncStatus   string     `json:"sync_status,omitempty"`
	Completeness float64    `json:"completeness,omitempty"`
}

// TradeSide 交易方向。
type TradeSide string

const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

// PortfolioTrade 交易记录（§4.6）。
type PortfolioTrade struct {
	TradeID   string    `json:"trade_id"`
	TradeDate time.Time `json:"trade_date"`
	StockCode string    `json:"stock_code"`
	StockName string    `json:"stock_name"`
	Side      TradeSide `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	TotalFee  float64   `json:"total_fee"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// PortfolioDividend 现金分红（§4.7）。
type PortfolioDividend struct {
	DividendID        string    `json:"dividend_id"`
	DividendDate      time.Time `json:"dividend_date"`
	StockCode         string    `json:"stock_code"`
	DividendPerShare  *float64  `json:"dividend_per_share,omitempty"`
	TotalDividend     float64   `json:"total_dividend"`
	BonusShareRatio   *float64  `json:"bonus_share_ratio,omitempty"`
	TransferShareRatio *float64 `json:"transfer_share_ratio,omitempty"`
	Note              string    `json:"note,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// CashFlowType 现金流类型。
type CashFlowType string

const (
	CashFlowDeposit    CashFlowType = "deposit"
	CashFlowWithdrawal CashFlowType = "withdrawal"
	CashFlowTrade      CashFlowType = "trade"
	CashFlowDividend   CashFlowType = "dividend"
)

// PortfolioCashFlow 现金流记录（§4.8）。
type PortfolioCashFlow struct {
	CashFlowID string       `json:"cash_flow_id"`
	FlowDate   time.Time    `json:"flow_date"`
	Amount     float64      `json:"amount"`
	FlowType   CashFlowType `json:"flow_type"`
	SourceRef  *string      `json:"source_ref,omitempty"`
	Note       string       `json:"note,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

// PortfolioPosition 持仓状态（§4.9）。
type PortfolioPosition struct {
	StockCode      string   `json:"stock_code"`
	StockName      string   `json:"stock_name"`
	Quantity       float64  `json:"quantity"`
	TotalCost      float64  `json:"total_cost"`
	AverageCost    float64  `json:"average_cost"`
	RealizedPnL    float64  `json:"realized_pnl"`
	DividendIncome float64  `json:"dividend_income"`
	LatestPrice    *float64 `json:"latest_price,omitempty"`
	MarketValue    *float64 `json:"market_value,omitempty"`
	UnrealizedPnL  *float64 `json:"unrealized_pnl,omitempty"`
}

// PortfolioValuation 估值快照（§4.10）。
type PortfolioValuation struct {
	ValuationID        string    `json:"valuation_id"`
	ValuationDate      time.Time `json:"valuation_date"`
	StockCode          *string   `json:"stock_code,omitempty"`
	Price              *float64  `json:"price,omitempty"`
	TotalAssetOverride *float64  `json:"total_asset_override,omitempty"`
	Source             string    `json:"source"`
	Note               string    `json:"note,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// AssetSnapshot 资产快照（§4.11）。
type AssetSnapshot struct {
	SnapshotDate       time.Time `json:"snapshot_date"`
	CashBalance        float64   `json:"cash_balance"`
	HoldingMarketValue float64   `json:"holding_market_value"`
	TotalAsset         float64   `json:"total_asset"`
	Source             string    `json:"source"`
}

// StockContribution 年度股票贡献。
type StockContribution struct {
	StockCode      string  `json:"stock_code"`
	RealizedPnL    float64 `json:"realized_pnl"`
	DividendIncome float64 `json:"dividend_income"`
}

// AnnualReview 年度复盘汇总（§4.12）。
type AnnualReview struct {
	Year             int                 `json:"year"`
	RealizedPnL      float64             `json:"realized_pnl"`
	DividendIncome   float64             `json:"dividend_income"`
	NetCashFlow      float64             `json:"net_cash_flow"`
	BeginTotalAsset  *float64            `json:"begin_total_asset,omitempty"`
	EndTotalAsset    *float64            `json:"end_total_asset,omitempty"`
	ReturnRate       *float64            `json:"return_rate,omitempty"`
	ByStockBreakdown []StockContribution `json:"by_stock_breakdown,omitempty"`
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
