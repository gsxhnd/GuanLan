"""Feature compute registry shared by training and serving."""

from __future__ import annotations

from collections.abc import Callable

import pandas as pd

from quant.ml.features.spec import FEATURE_COLUMNS, FEATURE_SPECS, FeatureSpec

FeatureFn = Callable[[pd.DataFrame], pd.Series]


def _return_1d(frame: pd.DataFrame) -> pd.Series:
    return frame["close"].pct_change()


def _close_ma5(frame: pd.DataFrame) -> pd.Series:
    return frame["close"].rolling(5, min_periods=1).mean()


def _close_ma20(frame: pd.DataFrame) -> pd.Series:
    return frame["close"].rolling(20, min_periods=1).mean()


def _volume_ma5(frame: pd.DataFrame) -> pd.Series:
    return frame["volume"].rolling(5, min_periods=1).mean()


def _volatility_20(frame: pd.DataFrame) -> pd.Series:
    return frame["close"].pct_change().rolling(20, min_periods=2).std()


_REGISTRY: dict[str, FeatureFn] = {
    "return_1d": _return_1d,
    "close_ma5": _close_ma5,
    "close_ma20": _close_ma20,
    "volume_ma5": _volume_ma5,
    "volatility_20": _volatility_20,
}


def list_features() -> list[str]:
    return list(FEATURE_COLUMNS)


def specs() -> tuple[FeatureSpec, ...]:
    return FEATURE_SPECS


def compute(frame: pd.DataFrame) -> pd.DataFrame:
    """Add registered feature columns. Expects sorted OHLCV with close/volume."""
    if frame.empty:
        out = frame.copy()
        for name in FEATURE_COLUMNS:
            out[name] = pd.Series(dtype="float64")
        return out
    out = frame.copy()
    for name in FEATURE_COLUMNS:
        fn = _REGISTRY[name]
        out[name] = fn(out)
    return out
