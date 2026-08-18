"""日频数据质量校验。"""

from __future__ import annotations

from dataclasses import dataclass
from datetime import date, timedelta
from typing import Iterable
from uuid import uuid4

import pandas as pd


@dataclass
class QualityIssue:
    stock_code: str
    trade_date: date | None
    issue_type: str
    severity: str
    message: str


def validate_bars(frame: pd.DataFrame, stock_code: str) -> list[QualityIssue]:
    issues: list[QualityIssue] = []
    if frame.empty:
        issues.append(
            QualityIssue(
                stock_code=stock_code,
                trade_date=None,
                issue_type="empty_dataset",
                severity="critical",
                message="no bars after cleaning",
            )
        )
        return issues

    for _, row in frame.iterrows():
        if row["volume"] == 0:
            issues.append(
                QualityIssue(
                    stock_code=stock_code,
                    trade_date=row["trade_date"],
                    issue_type="zero_volume",
                    severity="warning",
                    message="volume is zero",
                )
            )

    trade_dates = sorted(frame["trade_date"].tolist())
    gaps = _find_gaps(trade_dates, max_gap_days=7)
    for start, end in gaps:
        issues.append(
            QualityIssue(
                stock_code=stock_code,
                trade_date=start,
                issue_type="missing_range",
                severity="warning",
                message=f"missing trading days between {start} and {end}",
            )
        )

    return issues


def _find_gaps(dates: list[date], max_gap_days: int) -> list[tuple[date, date]]:
    gaps: list[tuple[date, date]] = []
    for prev, curr in zip(dates, dates[1:]):
        delta = (curr - prev).days
        if delta > max_gap_days:
            gaps.append((prev, curr))
    return gaps


def completeness_ratio(frame: pd.DataFrame, expected_days: int = 500) -> float:
    if frame.empty:
        return 0.0
    return min(100.0, round(len(frame) / expected_days * 100, 2))


def issue_rows(issues: Iterable[QualityIssue], data_version: str) -> list[dict]:
    from datetime import datetime, timezone

    now = datetime.now(timezone.utc)
    rows = []
    for issue in issues:
        rows.append(
            {
                "issue_id": str(uuid4()),
                "stock_code": issue.stock_code,
                "trade_date": issue.trade_date,
                "issue_type": issue.issue_type,
                "severity": issue.severity,
                "message": issue.message,
                "data_version": data_version,
                "created_at": now,
            }
        )
    return rows
