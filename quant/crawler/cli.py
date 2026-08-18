"""Crawler CLI (no --db / no scheduler)."""

from __future__ import annotations

import argparse
import json
import sys

from quant.crawler import pipeline


def cmd_fetch(args: argparse.Namespace) -> int:
    result = pipeline.fetch_stock(
        args.stock,
        lookback_days=None if args.full else args.lookback_days,
        full_history=args.full,
        start_date=args.start,
        end_date=args.end,
    )
    if result.error:
        print(result.error, file=sys.stderr)
        return 1
    payload = {
        "stock_code": result.stock_code,
        "bars": len(result.bars),
        "issues": [i.issue_type for i in result.issues],
        "completeness": pipeline.completeness_ratio(result.bars),
    }
    print(json.dumps(payload, ensure_ascii=False))
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(prog="quant.crawler", description="Stateless market data crawler")
    sub = parser.add_subparsers(dest="command", required=True)

    fetch = sub.add_parser("fetch", help="Fetch/clean one symbol (prints JSON summary)")
    fetch.add_argument("--stock", required=True)
    fetch.add_argument("--lookback-days", type=int, default=7)
    fetch.add_argument("--full", action="store_true")
    fetch.add_argument("--start", default=None)
    fetch.add_argument("--end", default=None)
    fetch.set_defaults(func=cmd_fetch)

    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
