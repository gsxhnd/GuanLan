"""Crawler gRPC server — stateless FetchDailyBars."""

from __future__ import annotations

from concurrent import futures

import grpc

from quant.core.logging import get_logger
from quant.crawler import pipeline
from quant.v1 import crawler_pb2, crawler_pb2_grpc

log = get_logger("quant.crawler.server")


def _bar_to_pb(row) -> crawler_pb2.CleanedDailyBar:
    trade_date = row["trade_date"]
    if hasattr(trade_date, "isoformat"):
        trade_date = trade_date.isoformat()
    return crawler_pb2.CleanedDailyBar(
        stock_code=str(row["stock_code"]),
        market=str(row.get("market") or ""),
        trade_date=str(trade_date),
        open=float(row["open"]),
        high=float(row["high"]),
        low=float(row["low"]),
        close=float(row["close"]),
        volume=int(row["volume"] or 0),
        source=str(row.get("source") or "yfinance"),
    )


class CrawlerService(crawler_pb2_grpc.CrawlerServiceServicer):
    def FetchDailyBars(self, request, context):  # noqa: N802
        codes = list(request.stock_codes) or []
        start = request.start_date or None
        end = request.end_date or None
        full = not start and not end

        for code in codes:
            result = pipeline.fetch_stock(
                code,
                start_date=start,
                end_date=end,
                full_history=full,
            )
            if result.error:
                yield crawler_pb2.FetchDailyBarsResponse(
                    stock_code=code,
                    error=result.error,
                )
                continue
            bars = [_bar_to_pb(row) for _, row in result.bars.iterrows()]
            yield crawler_pb2.FetchDailyBarsResponse(
                stock_code=code,
                bars=bars,
            )


def serve(addr: str = "[::]:50061") -> None:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    crawler_pb2_grpc.add_CrawlerServiceServicer_to_server(CrawlerService(), server)
    server.add_insecure_port(addr)
    server.start()
    log.info("crawler listening on %s", addr)
    server.wait_for_termination()


def main() -> None:
    import argparse

    parser = argparse.ArgumentParser(prog="quant.crawler.server")
    parser.add_argument("--addr", default="[::]:50061")
    args = parser.parse_args()
    serve(args.addr)


if __name__ == "__main__":
    main()
