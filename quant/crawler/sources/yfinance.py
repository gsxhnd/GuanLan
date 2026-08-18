"""使用 yfinance 获取 Yahoo Finance 日频 OHLCV 数据。"""

from __future__ import annotations

from datetime import date, datetime, timedelta, timezone
from typing import Any

import pandas as pd
import yfinance as yf

from quant.core.symbols import infer_market, to_yahoo_symbol


def _history_to_frame(stock_code: str, hist: pd.DataFrame) -> pd.DataFrame:
    if hist.empty:
        raise RuntimeError(f"yfinance returned no data for {stock_code}")

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
        frame[dst] = frame[src] if src in frame.columns else None

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


def fetch_daily_bars(
    stock_code: str,
    *,
    yfinance_symbol: str | None = None,
    period: str | None = None,
    lookback_days: int | None = None,
) -> pd.DataFrame:
    symbol = yfinance_symbol or to_yahoo_symbol(stock_code)
    ticker = yf.Ticker(symbol)

    if lookback_days is not None and lookback_days > 0:
        start = date.today() - timedelta(days=lookback_days + 5)
        end = date.today() + timedelta(days=1)
        hist = ticker.history(start=start.isoformat(), end=end.isoformat(), auto_adjust=False)
    else:
        hist = ticker.history(period=period or "2y", auto_adjust=False)

    return _history_to_frame(stock_code, hist)


def fetch_daily_bars_range(
    stock_code: str,
    *,
    yfinance_symbol: str | None = None,
    start: str | None = None,
    end: str | None = None,
) -> pd.DataFrame:
    symbol = yfinance_symbol or to_yahoo_symbol(stock_code)
    ticker = yf.Ticker(symbol)
    hist = ticker.history(start=start, end=end, auto_adjust=False)
    return _history_to_frame(stock_code, hist)


def fetch_ticker_info(stock_code: str, yfinance_symbol: str | None = None) -> dict[str, Any]:
    symbol = yfinance_symbol or to_yahoo_symbol(stock_code)
    info = yf.Ticker(symbol).get_info()
    return info if isinstance(info, dict) else {}
