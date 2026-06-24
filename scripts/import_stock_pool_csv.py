#!/usr/bin/env python3
"""将成分股 CSV 导入 DuckDB stock_pool。

与 daily_data 服务解耦，仅负责元数据入库。
"""

from __future__ import annotations

import argparse
import csv
import sys
from datetime import datetime, timezone
from pathlib import Path

import duckdb

SOURCE_CSV_IMPORT = "csv_import"

EXCHANGE_CN_TO_SUFFIX = {
    "上海证券交易所": ("SH", "SS"),
    "深圳证券交易所": ("SZ", "SZ"),
}


def ensure_stock_pool_schema(conn: duckdb.DuckDBPyConnection) -> None:
    legacy = conn.execute(
        """
        SELECT COUNT(*) FROM information_schema.columns
        WHERE table_name = 'stock_pool' AND column_name = 'stock_code'
        """
    ).fetchone()[0]
    if legacy:
        conn.execute("DROP TABLE stock_pool")

    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS stock_pool (
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
        )
        """
    )
    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_stock_pool_market ON stock_pool (market)"
    )
    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_stock_pool_source ON stock_pool (source)"
    )
    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_stock_pool_sync_daily ON stock_pool (sync_daily)"
    )


def parse_csi_constituent_csv(csv_path: Path) -> list[dict]:
    rows: list[dict] = []
    with csv_path.open(encoding="utf-8", newline="") as f:
        reader = csv.DictReader(f)
        for row in reader:
            original = row["成份券代码Constituent Code"].strip().zfill(6)
            name = row["成份券名称Constituent Name"].strip()
            exchange_cn = row["交易所Exchange"].strip()
            mapped = EXCHANGE_CN_TO_SUFFIX.get(exchange_cn)
            if not mapped:
                raise ValueError(f"unknown exchange: {exchange_cn!r} for {original}")
            exchange, yf_suffix = mapped
            yfinance_symbol = f"{original}.{yf_suffix}"
            rows.append(
                {
                    "yfinance_symbol": yfinance_symbol,
                    "original_code": original,
                    "market": "A",
                    "exchange": exchange,
                    "stock_name": name,
                    "currency": "CNY",
                    "source": SOURCE_CSV_IMPORT,
                }
            )
    return rows


def upsert_rows(conn: duckdb.DuckDBPyConnection, rows: list[dict]) -> int:
    now = datetime.now(timezone.utc)
    count = 0
    for row in rows:
        conn.execute(
            """
            INSERT INTO stock_pool (
                yfinance_symbol, original_code, market, exchange, stock_name, currency,
                source, is_active, sync_daily, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, TRUE, TRUE, ?, ?)
            ON CONFLICT (yfinance_symbol) DO UPDATE SET
                original_code = excluded.original_code,
                market = excluded.market,
                stock_name = COALESCE(NULLIF(excluded.stock_name, ''), stock_pool.stock_name),
                exchange = excluded.exchange,
                currency = excluded.currency,
                source = CASE
                    WHEN stock_pool.source = 'api_manual' THEN stock_pool.source
                    ELSE excluded.source
                END,
                updated_at = excluded.updated_at
            """,
            [
                row["yfinance_symbol"],
                row["original_code"],
                row["market"],
                row["exchange"],
                row["stock_name"],
                row["currency"],
                row["source"],
                now,
                now,
            ],
        )
        count += 1
    return count


def main() -> int:
    parser = argparse.ArgumentParser(description="Import constituent CSV into DuckDB stock_pool")
    parser.add_argument(
        "--csv",
        type=Path,
        default=Path("data/000510cons.csv"),
        help="CSI-style constituent CSV (default: A500)",
    )
    parser.add_argument(
        "--db",
        default="data/guanlan.duckdb",
        help="DuckDB database path",
    )
    parser.add_argument("--dry-run", action="store_true", help="Parse only, do not write")
    args = parser.parse_args()

    if not args.csv.is_file():
        print(f"csv not found: {args.csv}", file=sys.stderr)
        return 1

    rows = parse_csi_constituent_csv(args.csv)
    print(f"parsed {len(rows)} rows from {args.csv}")
    if args.dry_run:
        for row in rows[:3]:
            print(row)
        print("...")
        return 0

    db_path = Path(args.db)
    db_path.parent.mkdir(parents=True, exist_ok=True)
    conn = duckdb.connect(str(db_path))
    try:
        ensure_stock_pool_schema(conn)
        n = upsert_rows(conn, rows)
        print(f"imported {n} stocks into stock_pool (source={SOURCE_CSV_IMPORT})")
    finally:
        conn.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
