"""日频数据服务 CLI。

与模型训练服务 (services.training) 分离，仅负责行情拉取、清洗与入库。
股票池元数据请用 scripts/import_stock_pool_csv.py 导入。
"""

from __future__ import annotations

import argparse
import sys
import time
from pathlib import Path

from services.daily_data import db, pipeline


def cmd_sync(args: argparse.Namespace) -> int:
    conn = db.connect(args.db)
    try:
        version = pipeline.sync_stock(
            conn,
            args.stock,
            lookback_days=None if args.full else args.lookback_days,
            full_history=args.full,
        )
        print(version)
        return 0
    except Exception as exc:
        print(str(exc), file=sys.stderr)
        return 1
    finally:
        conn.close()


def cmd_daily_sync(args: argparse.Namespace) -> int:
    conn = db.connect(args.db)
    try:
        lines = pipeline.daily_sync(conn, lookback_days=args.lookback_days)
        for line in lines:
            print(line)
        print(f"synced {len(lines)} stocks")
        return 0
    except Exception as exc:
        print(str(exc), file=sys.stderr)
        return 1
    finally:
        conn.close()


def cmd_run_scheduler(args: argparse.Namespace) -> int:
    """独立日频服务：按 cron 间隔执行 daily-sync。"""
    interval = max(args.interval_hours, 1) * 3600
    print(f"daily_data scheduler started, interval={args.interval_hours}h, lookback={args.lookback_days}d")
    while True:
        conn = db.connect(args.db)
        try:
            lines = pipeline.daily_sync(conn, lookback_days=args.lookback_days)
            print(f"[daily-sync] {len(lines)} stocks")
        except Exception as exc:
            print(f"[daily-sync] error: {exc}", file=sys.stderr)
        finally:
            conn.close()
        time.sleep(interval)


def cmd_init_training(args: argparse.Namespace) -> int:
    indices = Path(args.indices) if args.indices else (
        Path(__file__).resolve().parents[2] / "services/daily_data/training/indices.json"
    )
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
    parser = argparse.ArgumentParser(
        prog="services.daily_data",
        description="GuanLan 日频行情数据服务（与 training 服务分离）",
    )
    parser.add_argument("--db", default="data/duckdb/guanlan.duckdb")
    sub = parser.add_subparsers(dest="command", required=True)

    sync = sub.add_parser("sync", help="同步单只股票（yfinance 符号）")
    sync.add_argument("--stock", required=True, help="如 600519.SS 或 000001.SZ")
    sync.add_argument("--lookback-days", type=int, default=7)
    sync.add_argument("--full", action="store_true", help="拉取完整历史")
    sync.set_defaults(func=cmd_sync)

    daily = sub.add_parser("daily-sync", help="增量同步 stock_pool 全部每日股票")
    daily.add_argument("--lookback-days", type=int, default=7)
    daily.set_defaults(func=cmd_daily_sync)

    sched = sub.add_parser("run-scheduler", help="以独立进程运行定时 daily-sync")
    sched.add_argument("--interval-hours", type=float, default=24)
    sched.add_argument("--lookback-days", type=int, default=7)
    sched.set_defaults(func=cmd_run_scheduler)

    init = sub.add_parser("init-training", help="(legacy) 按旧 indices.json 初始化")
    init.add_argument("--index-code", action="append")
    init.add_argument("--indices")
    init.set_defaults(func=cmd_init_training)

    args = parser.parse_args()
    if args.command == "run-scheduler":
        cmd_run_scheduler(args)
        return 0
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
