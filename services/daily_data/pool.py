"""数据底座股票池读写（元数据由 DuckDB 管理，CSV 导入见 scripts/import_stock_pool_csv.py）。"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

import duckdb

from services.daily_data.sources.yfinance import infer_market, to_yahoo_symbol

SOURCE_API_MANUAL = "api_manual"
SOURCE_CSV_IMPORT = "csv_import"


def upsert_stock_pool(conn: duckdb.DuckDBPyConnection, row: dict[str, Any]) -> None:
    now = datetime.now(timezone.utc)
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
            row.get("exchange"),
            row.get("stock_name", row["yfinance_symbol"]),
            row.get("currency", "USD" if row["market"] == "US" else "CNY"),
            row.get("source", SOURCE_API_MANUAL),
            now,
            now,
        ],
    )


def get_stock_pool_row(conn: duckdb.DuckDBPyConnection, yfinance_symbol: str) -> dict[str, Any] | None:
    row = conn.execute(
        """
        SELECT yfinance_symbol, original_code, market, exchange, stock_name, currency,
               source, is_active, sync_daily
        FROM stock_pool WHERE yfinance_symbol = ?
        """,
        [yfinance_symbol.strip()],
    ).fetchone()
    if not row:
        return None
    return {
        "yfinance_symbol": row[0],
        "original_code": row[1],
        "market": row[2],
        "exchange": row[3],
        "stock_name": row[4],
        "currency": row[5],
        "source": row[6],
        "is_active": row[7],
        "sync_daily": row[8],
    }


def list_daily_sync_stocks(conn: duckdb.DuckDBPyConnection) -> list[dict[str, Any]]:
    rows = conn.execute(
        """
        SELECT yfinance_symbol, original_code, market, stock_name
        FROM stock_pool
        WHERE is_active = TRUE AND sync_daily = TRUE
        ORDER BY yfinance_symbol
        """
    ).fetchall()
    return [
        {
            "yfinance_symbol": r[0],
            "original_code": r[1],
            "market": r[2],
            "stock_name": r[3],
        }
        for r in rows
    ]


def ensure_stock_in_pool(
    conn: duckdb.DuckDBPyConnection,
    symbol: str,
    market: str | None = None,
    stock_name: str | None = None,
) -> dict[str, Any]:
    code = symbol.strip().upper()
    yf = to_yahoo_symbol(code) if "." in code else code
    if "." in code and code.endswith((".SH", ".SZ")):
        original, suffix = code.rsplit(".", 1)
        exchange = suffix
        if suffix == "SH":
            yf = f"{original}.SS"
            exchange = "SH"
    elif yf.endswith(".SS"):
        original, exchange = yf[:-3], "SH"
    elif yf.endswith(".SZ"):
        original, exchange = yf[:-3], "SZ"
    else:
        original, exchange = yf, None

    existing = get_stock_pool_row(conn, yf)
    if existing:
        return existing

    m = market or infer_market(code)
    upsert_stock_pool(
        conn,
        {
            "yfinance_symbol": yf,
            "original_code": original,
            "market": m,
            "exchange": exchange,
            "stock_name": stock_name or yf,
            "currency": "CNY" if m == "A" else "USD",
            "source": SOURCE_API_MANUAL,
        },
    )
    return get_stock_pool_row(conn, yf) or {}
