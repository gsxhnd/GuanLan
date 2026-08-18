"""Feature registry unit tests (no network, no DuckDB file)."""

from __future__ import annotations

import pandas as pd

from quant.ml.features.registry import compute, list_features
from quant.ml.features.spec import FEATURE_COLUMNS


def test_feature_columns_stable() -> None:
    assert list_features() == FEATURE_COLUMNS
    assert "return_1d" in FEATURE_COLUMNS


def test_compute_features() -> None:
    frame = pd.DataFrame(
        {
            "close": [100.0, 102.0, 101.0, 103.0, 104.0],
            "volume": [10, 11, 12, 13, 14],
        }
    )
    out = compute(frame)
    for name in FEATURE_COLUMNS:
        assert name in out.columns
    assert pd.isna(out["return_1d"].iloc[0])
    assert abs(out["return_1d"].iloc[1] - 0.02) < 1e-9
    assert abs(out["close_ma5"].iloc[-1] - frame["close"].mean()) < 1e-9
