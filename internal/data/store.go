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
const DefaultDBPath = "data/guanlan.duckdb"

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

	return nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}
