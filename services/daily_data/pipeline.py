"""日频数据同步管线。"""

from __future__ import annotations

import json
from datetime import date, datetime, timezone
from pathlib import Path

from services.daily_data import clean, db, validate
from services.daily_data.sources import yfinance


def sync_stock(conn, stock_code: str, stock_name: str | None = None, training_index: str | None = None) -> str:
    code = stock_code.strip().upper()
    market = yfinance.infer_market(code)
    version = db.new_data_version(f"sync {code}")
    db.register_version(conn, version, f"daily sync for {code}")

    raw = yfinance.fetch_daily_bars(code)
    db.write_raw_bars(conn, raw)

    cleaned = clean.clean_bars(raw)
    issues = validate.validate_bars(cleaned, code)
    db.write_quality_issues(conn, validate.issue_rows(issues, version))

    db.write_standardized_bars(conn, cleaned, version)
    db.write_features(conn, cleaned, version)

    completeness = validate.completeness_ratio(cleaned)
    name = stock_name or code
    try:
        info = yfinance.fetch_ticker_info(code)
        name = info.get("shortName") or info.get("longName") or name
    except Exception:
        pass

    start = cleaned["trade_date"].min() if not cleaned.empty else None
    end = cleaned["trade_date"].max() if not cleaned.empty else None
    db.upsert_stock_status(
        conn,
        stock_code=code,
        stock_name=name,
        market=market,
        completeness=completeness,
        sync_status="ready",
        training_index_code=training_index,
        data_start=start,
        data_end=end,
    )
    return version


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
        db.upsert_index_constituent(conn, index_code, code, snap, item.get("weight"))
        v = sync_stock(conn, code, item.get("name"), training_index=index_code)
        row = conn.execute(
            "SELECT completeness FROM stock_data_status WHERE stock_code = ?", [code]
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
