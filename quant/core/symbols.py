"""Symbol normalization helpers (no I/O)."""

from __future__ import annotations


def to_yahoo_symbol(stock_code: str) -> str:
    code = stock_code.strip().upper()
    if "." not in code:
        return code
    sym, suffix = code.rsplit(".", 1)
    if suffix == "SH":
        return f"{sym}.SS"
    if suffix == "SZ":
        return f"{sym}.SZ"
    return code


def infer_market(stock_code: str) -> str:
    code = stock_code.strip().upper()
    if code.endswith(".SH") or code.endswith(".SS") or code.endswith(".SZ"):
        return "A"
    return "US"


def original_code(stock_code: str) -> str:
    code = stock_code.strip().upper()
    if "." in code:
        return code.split(".", 1)[0]
    return code
