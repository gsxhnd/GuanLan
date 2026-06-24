package data

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/duckdb/duckdb-go/v2"
)

// DefaultDBPath 默认 DuckDB 文件路径。
const DefaultDBPath = "data/duckdb/guanlan.duckdb"

// Store DuckDB 底层存储，承载数据页查询与写入。
type Store struct {
	db *sql.DB
}

// Open 打开（或创建）DuckDB 并执行 schema 迁移。
func Open(path string) (*Store, error) {
	if path == "" {
		path = DefaultDBPath
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping duckdb: %w", err)
	}

	s := &Store{db: db}
	if err := s.Migrate(context.Background()); err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

// OpenMemory 打开内存库，供测试使用。
func OpenMemory() (*Store, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open memory duckdb: %w", err)
	}
	s := &Store{db: db}
	if err := s.Migrate(context.Background()); err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

// DB 暴露底层连接，供批量写入等扩展场景使用。
func (s *Store) DB() *sql.DB {
	return s.db
}

// Migrate 创建数据页相关表结构。
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schemaDDL); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	if err := s.migrateStockPoolV2(ctx); err != nil {
		return err
	}
	return nil
}

// migrateStockPoolV2 将旧版 stock_pool（stock_code 主键）迁移为新结构。
func (s *Store) migrateStockPoolV2(ctx context.Context) error {
	var hasLegacy bool
	row := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_name = 'stock_pool' AND column_name = 'stock_code'
	`)
	if err := row.Scan(&hasLegacy); err != nil {
		return fmt.Errorf("inspect stock_pool schema: %w", err)
	}
	if !hasLegacy {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DROP TABLE stock_pool`); err != nil {
		return fmt.Errorf("drop legacy stock_pool: %w", err)
	}
	const stockPoolDDL = `
CREATE TABLE stock_pool (
	yfinance_symbol  VARCHAR PRIMARY KEY,
	original_code    VARCHAR NOT NULL,
	market           VARCHAR NOT NULL,
	exchange         VARCHAR,
	stock_name       VARCHAR NOT NULL DEFAULT '',
	currency         VARCHAR NOT NULL DEFAULT '',
	source           VARCHAR NOT NULL,
	is_active        BOOLEAN NOT NULL DEFAULT TRUE,
	sync_daily       BOOLEAN NOT NULL DEFAULT TRUE,
	created_at       TIMESTAMPTZ NOT NULL,
	updated_at       TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_stock_pool_market ON stock_pool (market);
CREATE INDEX IF NOT EXISTS idx_stock_pool_source ON stock_pool (source);
CREATE INDEX IF NOT EXISTS idx_stock_pool_sync_daily ON stock_pool (sync_daily);
`
	if _, err := s.db.ExecContext(ctx, stockPoolDDL); err != nil {
		return fmt.Errorf("create stock_pool v2: %w", err)
	}
	return nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}
