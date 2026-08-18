"""日频数据清洗。"""

from __future__ import annotations

import pandas as pd


def clean_bars(frame: pd.DataFrame) -> pd.DataFrame:
    if frame.empty:
        return frame

    out = frame.copy()
    numeric_cols = ["open", "high", "low", "close", "volume"]
    for col in numeric_cols:
        if col in out.columns:
            out[col] = pd.to_numeric(out[col], errors="coerce")

    out = out.dropna(subset=["open", "high", "low", "close", "trade_date"])
    out = out[out["volume"].fillna(0) >= 0]
    out = out[out["high"] >= out["low"]]
    out = out[(out["open"] >= out["low"]) & (out["open"] <= out["high"])]
    out = out[(out["close"] >= out["low"]) & (out["close"] <= out["high"])]
    out = out.drop_duplicates(subset=["stock_code", "trade_date"], keep="last")
    out = out.sort_values("trade_date")
    return out.reset_index(drop=True)
