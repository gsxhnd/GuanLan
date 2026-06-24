"""CLI: python -m services.daily_data sync|init-training"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

from services.daily_data import db, pipeline


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[2]


def cmd_sync(args: argparse.Namespace) -> int:
    conn = db.connect(args.db)
    try:
        version = pipeline.sync_stock(conn, args.stock)
        print(version)
        return 0
    except Exception as exc:
        print(str(exc), file=sys.stderr)
        return 1
    finally:
        conn.close()


def cmd_init_training(args: argparse.Namespace) -> int:
    indices = Path(args.indices) if args.indices else _repo_root() / "services/daily_data/training/indices.json"
    conn = db.connect(args.db)
    try:
        codes = args.index_code or ["000905.SH", "000510.SH"]
        for code in codes:
            version = pipeline.init_training_index(conn, code, indices)
            print(f"{code}\t{version}")
        return 0
    except Exception as exc:
        print(str(exc), file=sys.stderr)
        return 1
    finally:
        conn.close()


def main() -> int:
    parser = argparse.ArgumentParser(prog="services.daily_data")
    parser.add_argument("--db", default="data/duckdb/guanlan.duckdb")
    sub = parser.add_subparsers(dest="command", required=True)

    sync = sub.add_parser("sync", help="sync one stock into DuckDB")
    sync.add_argument("--stock", required=True)
    sync.set_defaults(func=cmd_sync)

    init = sub.add_parser("init-training", help="initialize training index constituents")
    init.add_argument("--index-code", action="append")
    init.add_argument("--indices")
    init.set_defaults(func=cmd_init_training)

    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
