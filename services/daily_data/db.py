"""DuckDB 读写。"""

from __future__ import annotations

from datetime import date, datetime, timezone
from pathlib import Path
from typing import Any
from uuid import uuid4

import duckdb
import pandas as pd


def ensure_schema(conn: duckdb.DuckDBPyConnection) -> None:
    legacy = conn.execute(
        """
        SELECT COUNT(*) FROM information_schema.columns
        WHERE table_name = 'stock_pool' AND column_name = 'stock_code'
        """
    ).fetchone()
    if legacy and legacy[0]:
        conn.execute("DROP TABLE stock_pool")

    exists = conn.execute(
        """
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'stock_pool'
        LIMIT 1
        """
    ).fetchone()
    if exists:
        return
    schema_path = Path(__file__).with_name("schema.sql")
    conn.execute(schema_path.read_text(encoding="utf-8"))


def connect(db_path: str) -> duckdb.DuckDBPyConnection:
    path = Path(db_path)
    path.parent.mkdir(parents=True, exist_ok=True)
    conn = duckdb.connect(str(path))
    ensure_schema(conn)
    return conn


def new_data_version(description: str) -> str:
    return f"v{datetime.now(timezone.utc).strftime('%Y%m%d_%H%M%S')}"


def register_version(conn: duckdb.DuckDBPyConnection, version_id: str, description: str) -> None:
    conn.execute(
        """
        INSERT INTO data_versions (version_id, description, created_at)
        VALUES (?, ?, ?)
        ON CONFLICT (version_id) DO NOTHING
        """,
        [version_id, description, datetime.now(timezone.utc)],
    )


def write_raw_bars(conn: duckdb.DuckDBPyConnection, frame: pd.DataFrame) -> None:
    if frame.empty:
        return
    conn.register("raw_df", frame)
    conn.execute(
        """
        INSERT INTO daily_bars_raw (
            stock_code, market, trade_date, open, high, low, close,
            volume, amount, source, ingested_at
        )
        SELECT stock_code, market, trade_date, open, high, low, close,
               CAST(volume AS BIGINT), amount, source, ingested_at
        FROM raw_df
        ON CONFLICT (stock_code, trade_date, source) DO UPDATE SET
            open = excluded.open,
            high = excluded.high,
            low = excluded.low,
            close = excluded.close,
            volume = excluded.volume,
            amount = excluded.amount,
            ingested_at = excluded.ingested_at
        """
    )
    conn.unregister("raw_df")


def write_standardized_bars(
    conn: duckdb.DuckDBPyConnection, frame: pd.DataFrame, data_version: str
) -> None:
    if frame.empty:
        return
    std = frame.copy()
    std["data_version"] = data_version
    std["adj_factor"] = None
    conn.register("std_df", std)
    conn.execute(
        """
        INSERT INTO daily_bars (
            stock_code, market, trade_date, open, high, low, close,
            volume, amount, adj_factor, source, data_version
        )
        SELECT stock_code, market, trade_date, open, high, low, close,
               CAST(volume AS BIGINT), amount, adj_factor, source, data_version
        FROM std_df
        ON CONFLICT (stock_code, trade_date) DO UPDATE SET
            market = excluded.market,
            open = excluded.open,
            high = excluded.high,
            low = excluded.low,
            close = excluded.close,
            volume = excluded.volume,
            amount = excluded.amount,
            source = excluded.source,
            data_version = excluded.data_version
        """
    )
    conn.unregister("std_df")


def write_features(conn: duckdb.DuckDBPyConnection, frame: pd.DataFrame, data_version: str) -> None:
    if frame.empty:
        return
    feat = frame.copy()
    feat["return_1d"] = feat["close"].pct_change()
    feat["close_ma5"] = feat["close"].rolling(5, min_periods=1).mean()
    feat["volume_ma5"] = feat["volume"].rolling(5, min_periods=1).mean()
    feat["data_version"] = data_version
    feat = feat[["stock_code", "trade_date", "return_1d", "volume_ma5", "close_ma5", "data_version"]]
    conn.register("feat_df", feat)
    conn.execute(
        """
        INSERT INTO daily_features (
            stock_code, trade_date, return_1d, volume_ma5, close_ma5, data_version
        )
        SELECT stock_code, trade_date, return_1d, volume_ma5, close_ma5, data_version
        FROM feat_df
        ON CONFLICT (stock_code, trade_date) DO UPDATE SET
            return_1d = excluded.return_1d,
            volume_ma5 = excluded.volume_ma5,
            close_ma5 = excluded.close_ma5,
            data_version = excluded.data_version
        """
    )
    conn.unregister("feat_df")


def write_quality_issues(conn: duckdb.DuckDBPyConnection, rows: list[dict[str, Any]]) -> None:
    if not rows:
        return
    df = pd.DataFrame(rows)
    conn.register("issue_df", df)
    conn.execute(
        """
        INSERT INTO data_quality_issues (
            issue_id, stock_code, trade_date, issue_type, severity,
            message, data_version, created_at
        )
        SELECT issue_id, stock_code, trade_date, issue_type, severity,
               message, data_version, created_at
        FROM issue_df
        """
    )
    conn.unregister("issue_df")


def upsert_stock_status(
    conn: duckdb.DuckDBPyConnection,
    stock_code: str,
    stock_name: str,
    market: str,
    completeness: float,
    sync_status: str,
    training_index_code: str | None = None,
    data_start: date | None = None,
    data_end: date | None = None,
) -> None:
    now = datetime.now(timezone.utc)
    conn.execute(
        """
        INSERT INTO stock_data_status (
            stock_code, stock_name, market, training_index_code,
            data_start_date, data_end_date, completeness, last_update, sync_status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT (stock_code) DO UPDATE SET
            stock_name = excluded.stock_name,
            market = excluded.market,
            training_index_code = COALESCE(excluded.training_index_code, stock_data_status.training_index_code),
            data_start_date = excluded.data_start_date,
            data_end_date = excluded.data_end_date,
            completeness = excluded.completeness,
            last_update = excluded.last_update,
            sync_status = excluded.sync_status
        """,
        [
            stock_code,
            stock_name,
            market,
            training_index_code,
            data_start,
            data_end,
            completeness,
            now,
            sync_status,
        ],
    )


def upsert_index_dataset(
    conn: duckdb.DuckDBPyConnection,
    index_code: str,
    market: str,
    index_name: str,
    completeness: float,
    sync_status: str,
) -> None:
    now = datetime.now(timezone.utc)
    conn.execute(
        """
        INSERT INTO index_datasets (
            index_code, market, index_name, data_completeness, last_sync_time, sync_status
        ) VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT (index_code) DO UPDATE SET
            market = excluded.market,
            index_name = excluded.index_name,
            data_completeness = excluded.data_completeness,
            last_sync_time = excluded.last_sync_time,
            sync_status = excluded.sync_status
        """,
        [index_code, market, index_name, completeness, now, sync_status],
    )


def upsert_index_constituent(
    conn: duckdb.DuckDBPyConnection,
    index_code: str,
    stock_code: str,
    snap_date: date,
    weight: float | None,
) -> None:
    conn.execute(
        """
        INSERT INTO index_constituents (index_code, stock_code, snap_date, weight, is_active)
        VALUES (?, ?, ?, ?, TRUE)
        ON CONFLICT (index_code, stock_code, snap_date) DO UPDATE SET
            weight = excluded.weight,
            is_active = TRUE
        """,
        [index_code, stock_code, snap_date, weight],
    )


def stock_has_ready_data(conn: duckdb.DuckDBPyConnection, stock_code: str) -> bool:
    row = conn.execute(
        """
        SELECT sync_status FROM stock_data_status WHERE stock_code = ?
        """,
        [stock_code],
    ).fetchone()
    return bool(row and row[0] == "ready")
