"""日频数据同步管线。"""

from __future__ import annotations

import json
from datetime import date
from pathlib import Path

from services.daily_data import clean, db, pool, validate
from services.daily_data.sources import yfinance


def sync_stock(
    conn,
    yfinance_symbol: str,
    stock_name: str | None = None,
    training_index: str | None = None,
    *,
    lookback_days: int | None = None,
    full_history: bool = False,
) -> str:
    symbol = yfinance_symbol.strip()
    pool_row = pool.get_stock_pool_row(conn, symbol)
    if pool_row:
        stock_name = stock_name or pool_row.get("stock_name")
        market = pool_row["market"]
    else:
        pool.ensure_stock_in_pool(conn, symbol, stock_name=stock_name)
        pool_row = pool.get_stock_pool_row(conn, symbol) or {}
        market = pool_row.get("market") or yfinance.infer_market(symbol)

    version = db.new_data_version(f"sync {symbol}")
    db.register_version(conn, version, f"daily sync for {symbol}")

    if full_history:
        raw = yfinance.fetch_daily_bars(symbol, yfinance_symbol=symbol, period="2y")
    else:
        days = lookback_days if lookback_days is not None else 7
        raw = yfinance.fetch_daily_bars(
            symbol, yfinance_symbol=symbol, lookback_days=days
        )

    db.write_raw_bars(conn, raw)
    cleaned = clean.clean_bars(raw)
    issues = validate.validate_bars(cleaned, symbol)
    db.write_quality_issues(conn, validate.issue_rows(issues, version))
    db.write_standardized_bars(conn, cleaned, version)
    db.write_features(conn, cleaned, version)

    completeness = validate.completeness_ratio(cleaned)
    name = stock_name or symbol
    try:
        info = yfinance.fetch_ticker_info(symbol, yfinance_symbol=symbol)
        name = info.get("shortName") or info.get("longName") or name
    except Exception:
        pass

    start = cleaned["trade_date"].min() if not cleaned.empty else None
    end = cleaned["trade_date"].max() if not cleaned.empty else None
    db.upsert_stock_status(
        conn,
        stock_code=symbol,
        stock_name=name,
        market=market,
        completeness=completeness,
        sync_status="ready",
        training_index_code=training_index,
        data_start=start,
        data_end=end,
    )
    return version


def daily_sync(conn, lookback_days: int = 7) -> list[str]:
    """对 stock_pool 中启用每日同步的股票做增量拉取（近 N 个交易日窗口）。"""
    stocks = pool.list_daily_sync_stocks(conn)
    versions: list[str] = []
    for row in stocks:
        v = sync_stock(
            conn,
            row["yfinance_symbol"],
            row.get("stock_name"),
            lookback_days=lookback_days,
            full_history=False,
        )
        versions.append(f"{row['yfinance_symbol']}\t{v}")
    return versions


def init_training_index(conn, index_code: str, indices_path: Path) -> str:
    payload = json.loads(indices_path.read_text(encoding="utf-8"))
    if index_code not in payload:
        raise ValueError(f"unknown training index: {index_code}")

    meta = payload[index_code]
    constituents = meta["constituents"]
    snap = date.today()
    version = db.new_data_version(f"training init {index_code}")
    db.register_version(conn, version, f"training data init for {index_code}")

    total_completeness = 0.0
    for item in constituents:
        code = item["stock_code"]
        yf = yfinance.to_yahoo_symbol(code)
        pool.ensure_stock_in_pool(conn, yf, meta.get("market", "A"), item.get("name"))
        db.upsert_index_constituent(conn, index_code, yf, snap, item.get("weight"))
        sync_stock(conn, yf, item.get("name"), training_index=index_code, full_history=True)
        row = conn.execute(
            "SELECT completeness FROM stock_data_status WHERE stock_code = ?", [yf]
        ).fetchone()
        total_completeness += float(row[0]) if row else 0.0

    avg = round(total_completeness / max(len(constituents), 1), 2)
    db.upsert_index_dataset(
        conn,
        index_code=index_code,
        market=meta.get("market", "A"),
        index_name=meta.get("index_name", index_code),
        completeness=avg,
        sync_status="ready",
    )
    return version
