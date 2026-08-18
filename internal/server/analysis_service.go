package server

import (
	"context"

	"github.com/gsxhnd/guanlan/internal/data"
	pb "github.com/gsxhnd/guanlan/internal/proto/quant/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Services) RunAnalysis(ctx context.Context, req *pb.RunAnalysisRequest) (*pb.Task, error) {
	task, err := s.Biz.Task.CreateAnalysisTask(
		ctx,
		req.GetTradeDate(),
		req.GetModelVersion(),
		req.GetStockCodes(),
		data.TriggerManual,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "run analysis: %v", err)
	}
	return s.toTask(task), nil
}

func (s *Services) PredictOnDemand(ctx context.Context, req *pb.PredictOnDemandRequest) (*pb.Prediction, error) {
	if s.Predict == nil {
		return nil, status.Errorf(codes.Unavailable, "prediction orchestrator not configured")
	}
	row, err := s.Predict.PredictOne(ctx, req.GetStockCode(), req.GetTradeDate(), req.GetModelVersion())
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "predict: %v", err)
	}
	return toPredictionPB(row), nil
}

func (s *Services) ListPredictions(ctx context.Context, req *pb.ListPredictionsRequest) (*pb.ListPredictionsResponse, error) {
	rows, err := s.Store.ListPredictions(ctx, req.GetStockCode(), req.GetTradeDate(), req.GetModelVersion(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list predictions: %v", err)
	}
	out := make([]*pb.Prediction, 0, len(rows))
	for _, row := range rows {
		out = append(out, toPredictionPB(row))
	}
	return &pb.ListPredictionsResponse{Predictions: out}, nil
}

func toPredictionPB(p data.Prediction) *pb.Prediction {
	return &pb.Prediction{
		PredictionId: p.PredictionID,
		StockCode:    p.StockCode,
		TradeDate:    dateStr(p.TradeDate),
		Score:        p.Score,
		ModelVersion: p.ModelVersion,
		CreatedAt:    timestamppb.New(p.CreatedAt),
	}
}
