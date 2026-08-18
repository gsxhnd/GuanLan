package biz

import "testing"

func TestEncodeDecodeAnalysisTarget(t *testing.T) {
	enc := EncodeAnalysisTarget("2026-08-17", "baseline", []string{"aapl", " 600519.ss "})
	date, model, codes := DecodeAnalysisTarget(enc)
	if date != "2026-08-17" || model != "baseline" {
		t.Fatalf("got date=%s model=%s", date, model)
	}
	if len(codes) != 2 || codes[0] != "AAPL" || codes[1] != "600519.SS" {
		t.Fatalf("codes=%v", codes)
	}

	enc = EncodeAnalysisTarget("", "", nil)
	_, model, codes = DecodeAnalysisTarget(enc)
	if model != "latest" || len(codes) != 1 || codes[0] != AnalysisTargetWatchlist {
		t.Fatalf("watchlist target: model=%s codes=%v", model, codes)
	}
}
