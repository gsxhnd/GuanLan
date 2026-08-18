"""Prediction gRPC server — lazy-loads a baseline JSON artifact; no torch at import time."""

from __future__ import annotations

import json
import zlib
from concurrent import futures
from pathlib import Path

import grpc
from grpc import StatusCode

from quant.core.config import repo_root
from quant.core.logging import get_logger
from quant.v1 import prediction_pb2, prediction_pb2_grpc

log = get_logger("quant.ml.serving")

_torch_mod = None


def _try_import_torch():
    """Lazy torch import; unused until a torch artifact exists."""
    global _torch_mod
    if _torch_mod is not None:
        return _torch_mod
    try:
        import torch  # type: ignore

        _torch_mod = torch
        return torch
    except Exception:
        return None


class ModelStore:
    def __init__(self, models_dir: Path) -> None:
        self.models_dir = models_dir
        self._cache: dict[str, dict] = {}

    def latest_version(self) -> str | None:
        files = sorted(self.models_dir.glob("*.json"))
        if not files:
            return None
        return files[-1].stem

    def get(self, version: str) -> dict | None:
        if version in self._cache:
            return self._cache[version]
        path = self.models_dir / f"{version}.json"
        if not path.is_file():
            return None
        artifact = json.loads(path.read_text(encoding="utf-8"))
        self._cache[version] = artifact
        kind = artifact.get("kind")
        if kind == "torch":
            _try_import_torch()
        log.info("loaded model %s from %s", version, path)
        return artifact


def _score(stock_code: str, artifact: dict) -> float:
    intercept = float(artifact.get("intercept", 0.5))
    # Stable per-symbol offset so a baseline run is visually distinguishable.
    h = zlib.crc32(stock_code.encode("utf-8")) % 1000 / 1000.0
    score = intercept + (h - 0.5) * 0.2
    return max(0.0, min(1.0, score))


class PredictionService(prediction_pb2_grpc.PredictionServiceServicer):
    def __init__(self, models_dir: Path) -> None:
        self.store = ModelStore(models_dir)

    def _resolve(self, requested: str, context) -> dict | None:
        version = (requested or "").strip()
        if not version or version == "latest":
            version = self.store.latest_version() or "baseline"
        artifact = self.store.get(version)
        if artifact is None:
            context.set_code(StatusCode.UNIMPLEMENTED)
            context.set_details(f"model {version!r} not loaded; run quant.ml.training")
            return None
        return artifact

    def Predict(self, request, context):  # noqa: N802
        artifact = self._resolve(request.model_version, context)
        if artifact is None:
            return prediction_pb2.PredictResponse()
        code = request.stock_code.strip()
        version = artifact.get("model_version", "baseline")
        return prediction_pb2.PredictResponse(
            stock_code=code,
            trade_date=request.trade_date,
            score=_score(code, artifact),
            model_version=version,
        )

    def PredictBatch(self, request, context):  # noqa: N802
        artifact = self._resolve(request.model_version, context)
        if artifact is None:
            return prediction_pb2.PredictBatchResponse()
        version = artifact.get("model_version", "baseline")
        out = []
        for code in request.stock_codes:
            code = code.strip()
            if not code:
                continue
            out.append(
                prediction_pb2.PredictResponse(
                    stock_code=code,
                    trade_date=request.trade_date,
                    score=_score(code, artifact),
                    model_version=version,
                )
            )
        return prediction_pb2.PredictBatchResponse(predictions=out)


def serve(addr: str = "[::]:50062", models_dir: str = "data/models") -> None:
    path = Path(models_dir)
    if not path.is_absolute():
        path = repo_root() / path
    path.mkdir(parents=True, exist_ok=True)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=2))
    prediction_pb2_grpc.add_PredictionServiceServicer_to_server(
        PredictionService(path), server
    )
    server.add_insecure_port(addr)
    server.start()
    log.info("prediction listening on %s (models=%s)", addr, path)
    server.wait_for_termination()


def main() -> None:
    import argparse

    parser = argparse.ArgumentParser(prog="quant.ml.serving")
    parser.add_argument("--addr", default="[::]:50062")
    parser.add_argument("--models", default="data/models")
    args = parser.parse_args()
    serve(args.addr, args.models)


if __name__ == "__main__":
    main()
