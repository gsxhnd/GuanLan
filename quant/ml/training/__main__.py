"""Training CLI: read a Parquet snapshot in-memory and write a baseline model artifact.

Never opens data/guanlan.duckdb. Torch is not imported (lazy path is serving-only).
"""

from __future__ import annotations

import argparse
import json
import sys
from datetime import datetime, timezone
from pathlib import Path

from quant.core.config import repo_root
from quant.ml.features.registry import compute, list_features


def _models_dir(raw: str) -> Path:
    path = Path(raw)
    if not path.is_absolute():
        path = repo_root() / path
    path.mkdir(parents=True, exist_ok=True)
    return path


def _fit_intercept(snapshot: str | None) -> tuple[float, int]:
    """Optional snapshot fit; always falls back to intercept=0.5."""
    if not snapshot:
        return 0.5, 0
    snap = Path(snapshot)
    if not snap.is_file():
        raise FileNotFoundError(f"snapshot not found: {snap}")

    import duckdb  # in-memory read of parquet only

    con = duckdb.connect(":memory:")
    try:
        bars = con.execute(
            "SELECT stock_code, trade_date, open, high, low, close, volume "
            f"FROM read_parquet('{snap.as_posix()}') "
            "ORDER BY stock_code, trade_date"
        ).fetchdf()
    finally:
        con.close()

    if bars.empty:
        return 0.5, 0
    grouped = []
    for _, grp in bars.groupby("stock_code", sort=False):
        grouped.append(compute(grp.reset_index(drop=True)))
    import pandas as pd

    feats = pd.concat(grouped, ignore_index=True)
    col = "return_1d"
    if col in feats.columns and feats[col].notna().any():
        intercept = float(0.5 + feats[col].median())
        intercept = max(0.0, min(1.0, intercept))
        return intercept, int(len(feats))
    return 0.5, int(len(feats))


def main() -> int:
    parser = argparse.ArgumentParser(prog="quant.ml.training")
    parser.add_argument("--model", default="baseline")
    parser.add_argument("--snapshot", default="", help="Parquet snapshot path (optional)")
    parser.add_argument("--out", default="data/models", help="Model artifact directory")
    args = parser.parse_args()

    snapshot = args.snapshot or None
    intercept, rows = _fit_intercept(snapshot)
    version = args.model
    artifact = {
        "model_version": version,
        "kind": "intercept",
        "intercept": intercept,
        "feature_columns": list_features(),
        "rows_fitted": rows,
        "created_at": datetime.now(timezone.utc).isoformat(),
        "snapshot": snapshot or "",
    }
    out_dir = _models_dir(args.out)
    path = out_dir / f"{version}.json"
    path.write_text(json.dumps(artifact, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps({"model_version": version, "artifact": str(path), "rows_fitted": rows}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
