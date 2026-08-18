"""Shared feature column spec — used by training and serving to avoid train/serve skew."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class FeatureSpec:
    name: str
    description: str


FEATURE_SPECS: tuple[FeatureSpec, ...] = (
    FeatureSpec("return_1d", "Close-to-close 1-day return"),
    FeatureSpec("close_ma5", "5-day simple moving average of close"),
    FeatureSpec("close_ma20", "20-day simple moving average of close"),
    FeatureSpec("volume_ma5", "5-day simple moving average of volume"),
    FeatureSpec("volatility_20", "20-day rolling std of 1-day return"),
)

FEATURE_COLUMNS: list[str] = [spec.name for spec in FEATURE_SPECS]
