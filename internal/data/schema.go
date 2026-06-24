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

CREATE TABLE IF NOT EXISTS data_sync_tasks (
	task_id        VARCHAR PRIMARY KEY,
	task_type      VARCHAR NOT NULL DEFAULT 'data_sync',
	target_object  VARCHAR NOT NULL,
	trigger_method VARCHAR NOT NULL,
	status         VARCHAR NOT NULL DEFAULT 'pending',
	created_at     TIMESTAMPTZ NOT NULL,
	started_at     TIMESTAMPTZ,
	ended_at       TIMESTAMPTZ,
	retry_count    INTEGER NOT NULL DEFAULT 0,
	failure_reason VARCHAR,
	log_ref        VARCHAR
);

CREATE INDEX IF NOT EXISTS idx_data_sync_tasks_created_at
	ON data_sync_tasks (created_at DESC);
`
