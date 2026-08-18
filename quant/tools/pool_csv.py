"""Import CSI constituent CSV via api HTTP (no DuckDB)."""

from __future__ import annotations

import argparse
import csv
import json
import sys
import urllib.error
import urllib.request
from pathlib import Path

from quant.core.config import default_api_base

SOURCE_CSV_IMPORT = "csv_import"

EXCHANGE_CN_TO_SUFFIX = {
    "上海证券交易所": ("SH", "SS"),
    "深圳证券交易所": ("SZ", "SZ"),
}


def parse_csi_constituent_csv(csv_path: Path) -> list[dict]:
    rows: list[dict] = []
    with csv_path.open(encoding="utf-8", newline="") as f:
        reader = csv.DictReader(f)
        for row in reader:
            original = row["成份券代码Constituent Code"].strip().zfill(6)
            name = row["成份券名称Constituent Name"].strip()
            exchange_cn = row["交易所Exchange"].strip()
            mapped = EXCHANGE_CN_TO_SUFFIX.get(exchange_cn)
            if not mapped:
                raise ValueError(f"unknown exchange: {exchange_cn!r} for {original}")
            exchange, yf_suffix = mapped
            yfinance_symbol = f"{original}.{yf_suffix}"
            rows.append(
                {
                    "yfinanceSymbol": yfinance_symbol,
                    "originalCode": original,
                    "market": "A",
                    "exchange": exchange,
                    "stockName": name,
                    "currency": "CNY",
                    "syncDaily": True,
                    "isActive": True,
                }
            )
    return rows


def upsert_via_api(api_base: str, row: dict) -> None:
    url = f"{api_base.rstrip('/')}/api/data/pool/items"
    body = json.dumps(row).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        if resp.status >= 300:
            raise RuntimeError(f"api {resp.status}: {resp.read().decode()}")


def main() -> int:
    parser = argparse.ArgumentParser(prog="quant.tools.pool_csv")
    parser.add_argument("--csv", required=True, type=Path)
    parser.add_argument("--api", default=default_api_base())
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    rows = parse_csi_constituent_csv(args.csv)
    print(f"parsed {len(rows)} rows from {args.csv}")
    if args.dry_run:
        for row in rows[:5]:
            print(json.dumps(row, ensure_ascii=False))
        return 0

    ok = 0
    for row in rows:
        try:
            upsert_via_api(args.api, row)
            ok += 1
        except urllib.error.URLError as exc:
            print(f"fail {row['yfinanceSymbol']}: {exc}", file=sys.stderr)
        except Exception as exc:  # noqa: BLE001
            print(f"fail {row['yfinanceSymbol']}: {exc}", file=sys.stderr)
    print(f"upserted {ok}/{len(rows)} via {args.api}")
    return 0 if ok == len(rows) else 1


if __name__ == "__main__":
    raise SystemExit(main())
