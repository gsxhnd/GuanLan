package data

// schemaDDL 创建数据页相关底层表。命名与 docs/dev/03-domain-model.md 字段对齐。
const schemaDDL = `
CREATE TABLE IF NOT EXISTS index_datasets (
	index_code        VARCHAR PRIMARY KEY,
	market            VARCHAR NOT NULL,
	index_name        VARCHAR NOT NULL,
	data_completeness DOUBLE NOT NULL DEFAULT 0,
	last_sync_time    TIMESTAMPTZ,
	sync_status       VARCHAR NOT NULL DEFAULT 'pending'
);

CREATE TABLE IF NOT EXISTS index_constituents (
	index_code  VARCHAR NOT NULL,
	stock_code  VARCHAR NOT NULL,
	snap_date   DATE NOT NULL,
	weight      DOUBLE,
	is_active   BOOLEAN NOT NULL DEFAULT TRUE,
	PRIMARY KEY (index_code, stock_code, snap_date)
);

CREATE TABLE IF NOT EXISTS daily_bars_raw (
	stock_code   VARCHAR NOT NULL,
	market       VARCHAR NOT NULL,
	trade_date   DATE NOT NULL,
	open         DOUBLE,
	high         DOUBLE,
	low          DOUBLE,
	close        DOUBLE,
	volume       BIGINT,
	amount       DOUBLE,
	source       VARCHAR NOT NULL DEFAULT 'yfinance',
	ingested_at  TIMESTAMPTZ NOT NULL,
	PRIMARY KEY (stock_code, trade_date, source)
);

CREATE TABLE IF NOT EXISTS daily_bars (
	stock_code   VARCHAR NOT NULL,
	market       VARCHAR NOT NULL,
	trade_date   DATE NOT NULL,
	open         DOUBLE NOT NULL,
	high         DOUBLE NOT NULL,
	low          DOUBLE NOT NULL,
	close        DOUBLE NOT NULL,
	volume       BIGINT NOT NULL,
	amount       DOUBLE,
	adj_factor   DOUBLE,
	source       VARCHAR NOT NULL DEFAULT '',
	data_version VARCHAR NOT NULL DEFAULT '',
	PRIMARY KEY (stock_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_bars_stock_date
	ON daily_bars (stock_code, trade_date DESC);

CREATE TABLE IF NOT EXISTS daily_features (
	stock_code    VARCHAR NOT NULL,
	trade_date    DATE NOT NULL,
	return_1d     DOUBLE,
	volume_ma5    DOUBLE,
	close_ma5     DOUBLE,
	data_version  VARCHAR NOT NULL DEFAULT '',
	PRIMARY KEY (stock_code, trade_date)
);

CREATE TABLE IF NOT EXISTS data_quality_issues (
	issue_id      VARCHAR PRIMARY KEY,
	stock_code    VARCHAR NOT NULL,
	trade_date    DATE,
	issue_type    VARCHAR NOT NULL,
	severity      VARCHAR NOT NULL,
	message       VARCHAR,
	data_version  VARCHAR,
	created_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_data_quality_stock
	ON data_quality_issues (stock_code, created_at DESC);

CREATE TABLE IF NOT EXISTS stock_data_status (
	stock_code          VARCHAR PRIMARY KEY,
	stock_name          VARCHAR NOT NULL DEFAULT '',
	market              VARCHAR NOT NULL,
	training_index_code VARCHAR,
	data_start_date     DATE,
	data_end_date       DATE,
	completeness        DOUBLE NOT NULL DEFAULT 0,
	missing_ranges      JSON,
	last_update         TIMESTAMPTZ,
	sync_status         VARCHAR NOT NULL DEFAULT 'missing'
);

CREATE INDEX IF NOT EXISTS idx_stock_data_status_market
	ON stock_data_status (market);

CREATE INDEX IF NOT EXISTS idx_stock_data_status_sync_status
	ON stock_data_status (sync_status);

CREATE TABLE IF NOT EXISTS tasks (
	task_id        VARCHAR PRIMARY KEY,
	task_type      VARCHAR NOT NULL,
	target_object  VARCHAR NOT NULL,
	trigger_method VARCHAR NOT NULL,
	status         VARCHAR NOT NULL DEFAULT 'pending',
	created_at     TIMESTAMPTZ NOT NULL,
	started_at     TIMESTAMPTZ,
	ended_at       TIMESTAMPTZ,
	retry_count    INTEGER NOT NULL DEFAULT 0,
	failure_reason VARCHAR,
	log_ref        VARCHAR,
	data_version   VARCHAR
);

CREATE INDEX IF NOT EXISTS idx_tasks_created_at
	ON tasks (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tasks_status
	ON tasks (status);

CREATE TABLE IF NOT EXISTS data_versions (
	version_id   VARCHAR PRIMARY KEY,
	description  VARCHAR,
	created_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS watchlist_items (
	stock_code     VARCHAR PRIMARY KEY,
	market         VARCHAR NOT NULL,
	tags           JSON,
	priority       INTEGER NOT NULL DEFAULT 0,
	notes          VARCHAR,
	is_active      BOOLEAN NOT NULL DEFAULT TRUE,
	added_at       TIMESTAMPTZ NOT NULL,
	removed_at     TIMESTAMPTZ,
	source         VARCHAR NOT NULL DEFAULT 'manual',
	last_action    VARCHAR,
	last_action_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_watchlist_items_active
	ON watchlist_items (is_active);

CREATE TABLE IF NOT EXISTS portfolio_trades (
	trade_id    VARCHAR PRIMARY KEY,
	trade_date  DATE NOT NULL,
	stock_code  VARCHAR NOT NULL,
	stock_name  VARCHAR NOT NULL DEFAULT '',
	side        VARCHAR NOT NULL,
	price       DOUBLE NOT NULL,
	quantity    DOUBLE NOT NULL,
	total_fee   DOUBLE NOT NULL DEFAULT 0,
	note        VARCHAR,
	created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_portfolio_trades_date
	ON portfolio_trades (trade_date DESC);

CREATE TABLE IF NOT EXISTS portfolio_dividends (
	dividend_id          VARCHAR PRIMARY KEY,
	dividend_date        DATE NOT NULL,
	stock_code           VARCHAR NOT NULL,
	dividend_per_share   DOUBLE,
	total_dividend       DOUBLE NOT NULL,
	bonus_share_ratio    DOUBLE,
	transfer_share_ratio DOUBLE,
	note                 VARCHAR,
	created_at           TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_portfolio_dividends_date
	ON portfolio_dividends (dividend_date DESC);

CREATE TABLE IF NOT EXISTS portfolio_cash_flows (
	cash_flow_id VARCHAR PRIMARY KEY,
	flow_date    DATE NOT NULL,
	amount       DOUBLE NOT NULL,
	flow_type    VARCHAR NOT NULL,
	source_ref   VARCHAR,
	note         VARCHAR,
	created_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_portfolio_cash_flows_date
	ON portfolio_cash_flows (flow_date DESC);

CREATE TABLE IF NOT EXISTS portfolio_valuations (
	valuation_id         VARCHAR PRIMARY KEY,
	valuation_date       DATE NOT NULL,
	stock_code           VARCHAR,
	price                DOUBLE,
	total_asset_override DOUBLE,
	source               VARCHAR NOT NULL DEFAULT 'manual',
	note                 VARCHAR,
	created_at           TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_portfolio_valuations_date
	ON portfolio_valuations (valuation_date DESC);

CREATE TABLE IF NOT EXISTS portfolio_asset_snapshots (
	snapshot_date          DATE PRIMARY KEY,
	cash_balance           DOUBLE NOT NULL,
	holding_market_value   DOUBLE NOT NULL,
	total_asset            DOUBLE NOT NULL,
	source                 VARCHAR NOT NULL
);
`
