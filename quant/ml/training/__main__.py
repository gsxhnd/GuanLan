"""模型训练服务 CLI — 读取 DuckDB 特征与标签，训练并导出模型版本。

行情获取请使用: python -m services.daily_data
"""

from __future__ import annotations

import argparse
import sys


def cmd_train(args: argparse.Namespace) -> int:
    print(
        f"training stub: model={args.model} data_version={args.data_version}",
        file=sys.stderr,
    )
    print("training service not implemented yet; use daily_data for market data")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(prog="services.training")
    parser.add_argument("--db", default="data/duckdb/guanlan.duckdb")
    sub = parser.add_subparsers(dest="command", required=True)

    train = sub.add_parser("train", help="训练模型（占位）")
    train.add_argument("--model", default="gru")
    train.add_argument("--data-version", default="latest")
    train.set_defaults(func=cmd_train)

    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
