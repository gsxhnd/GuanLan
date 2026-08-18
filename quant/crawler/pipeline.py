"""Stateless crawler pipeline: fetch → clean → validate (no DB)."""

from __future__ import annotations

from dataclasses import dataclass
from datetime import date

import pandas as pd

from quant.crawler import clean, validate
from quant.crawler.sources import yfinance


@dataclass
class FetchResult:
    stock_code: str
    bars: pd.DataFrame
    issues: list[validate.QualityIssue]
    error: str | None = None


def fetch_stock(
    stock_code: str,
    *,
    start_date: str | None = None,
    end_date: str | None = None,
    lookback_days: int | None = None,
    full_history: bool = False,
) -> FetchResult:
    """Pure fetch/clean/validate for one symbol. Never opens DuckDB."""
    symbol = stock_code.strip()
    try:
        if full_history or (not start_date and not lookback_days):
            raw = yfinance.fetch_daily_bars(symbol, yfinance_symbol=symbol, period="2y")
        elif start_date:
            raw = yfinance.fetch_daily_bars_range(
                symbol,
                yfinance_symbol=symbol,
                start=start_date,
                end=end_date,
            )
        else:
            days = lookback_days if lookback_days is not None else 7
            raw = yfinance.fetch_daily_bars(
                symbol, yfinance_symbol=symbol, lookback_days=days
            )
        cleaned = clean.clean_bars(raw)
        issues = validate.validate_bars(cleaned, symbol)
        return FetchResult(stock_code=symbol, bars=cleaned, issues=issues)
    except Exception as exc:  # noqa: BLE001 — surface to gRPC caller
        return FetchResult(
            stock_code=symbol,
            bars=pd.DataFrame(),
            issues=[],
            error=str(exc),
        )


def completeness_ratio(frame: pd.DataFrame, expected_days: int = 500) -> float:
    return validate.completeness_ratio(frame, expected_days=expected_days)
