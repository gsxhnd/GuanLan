"""使用 yfinance 获取 Yahoo Finance 日频 OHLCV 数据。"""

from __future__ import annotations

import yfinance as yf


if __name__ == "__main__":
    dat = yf.Ticker("000333.SZ")
    print(dat.get_info())
