"""Shared config helpers."""

from __future__ import annotations

import os
from pathlib import Path


def repo_root() -> Path:
    return Path(__file__).resolve().parents[2]


def default_api_base() -> str:
    return os.environ.get("GUANLAN_API_BASE", "http://127.0.0.1:8080")
