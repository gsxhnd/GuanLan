"""Python 日频数据管线单元测试（无网络）。"""

from __future__ import annotations

import pandas as pd

from quant.crawler import clean, validate
from quant.core.symbols import infer_market, to_yahoo_symbol


def test_to_yahoo_symbol() -> None:
    assert to_yahoo_symbol("600519.SH") == "600519.SS"
    assert to_yahoo_symbol("000858.SZ") == "000858.SZ"
    assert to_yahoo_symbol("AAPL") == "AAPL"


def test_infer_market() -> None:
    assert infer_market("600519.SH") == "A"
    assert infer_market("AAPL") == "US"


def test_clean_and_validate() -> None:
    frame = pd.DataFrame(
        {
            "stock_code": ["AAPL", "AAPL"],
            "market": ["US", "US"],
            "trade_date": [pd.Timestamp("2026-01-02").date(), pd.Timestamp("2026-01-03").date()],
            "open": [100.0, 101.0],
            "high": [102.0, 103.0],
            "low": [99.0, 100.0],
            "close": [101.0, 102.0],
            "volume": [1000, 0],
            "amount": [None, None],
            "source": ["yfinance", "yfinance"],
            "ingested_at": [pd.Timestamp.utcnow(), pd.Timestamp.utcnow()],
        }
    )
    cleaned = clean.clean_bars(frame)
    assert len(cleaned) == 2
    issues = validate.validate_bars(cleaned, "AAPL")
    assert any(i.issue_type == "zero_volume" for i in issues)
    assert validate.completeness_ratio(cleaned, expected_days=2) == 100.0
