"""使用 yfinance 获取 Yahoo Finance 日频 OHLCV 数据。"""

from __future__ import annotations

from datetime import datetime, timedelta, timezone
from typing import Any

import pandas as pd
import yfinance as yf


def to_yahoo_symbol(stock_code: str) -> str:
    code = stock_code.strip().upper()
    if "." not in code:
        return code
    sym, suffix = code.rsplit(".", 1)
    if suffix == "SH":
        return f"{sym}.SS"
    if suffix == "SZ":
        return f"{sym}.SZ"
    return code


def infer_market(stock_code: str) -> str:
    code = stock_code.strip().upper()
    if code.endswith(".SH") or code.endswith(".SZ"):
        return "A"
    return "US"


def fetch_daily_bars(stock_code: str, period: str = "2y") -> pd.DataFrame:
    symbol = to_yahoo_symbol(stock_code)
    ticker = yf.Ticker(symbol)
    hist = ticker.history(period=period, auto_adjust=False)
    if hist.empty:
        raise RuntimeError(f"yfinance returned no data for {stock_code} ({symbol})")

    frame = hist.reset_index()
    if "Date" in frame.columns:
        frame = frame.rename(columns={"Date": "trade_date"})
    elif "Datetime" in frame.columns:
        frame = frame.rename(columns={"Datetime": "trade_date"})
    else:
        raise RuntimeError("unexpected yfinance columns")

    frame["trade_date"] = pd.to_datetime(frame["trade_date"], utc=True).dt.date
    frame["stock_code"] = stock_code.strip().upper()
    frame["market"] = infer_market(stock_code)
    frame["source"] = "yfinance"
    frame["ingested_at"] = datetime.now(timezone.utc)

    cols = {
        "Open": "open",
        "High": "high",
        "Low": "low",
        "Close": "close",
        "Volume": "volume",
    }
    for src, dst in cols.items():
        if src in frame.columns:
            frame[dst] = frame[src]
        else:
            frame[dst] = None

    if "amount" not in frame.columns:
        frame["amount"] = None

    return frame[
        [
            "stock_code",
            "market",
            "trade_date",
            "open",
            "high",
            "low",
            "close",
            "volume",
            "amount",
            "source",
            "ingested_at",
        ]
    ]


def fetch_ticker_info(stock_code: str) -> dict[str, Any]:
    symbol = to_yahoo_symbol(stock_code)
    info = yf.Ticker(symbol).get_info()
    return info if isinstance(info, dict) else {}
