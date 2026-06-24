package server

import (
	"context"

	"github.com/gsxhnd/guanlan/internal/data"
	pb "github.com/gsxhnd/guanlan/internal/proto/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Services) ListIndexes(ctx context.Context, _ *pb.Empty) (*pb.ListIndexesResponse, error) {
	indexes, err := s.Store.ListIndexDatasets(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list indexes: %v", err)
	}
	out := make([]*pb.IndexDataset, 0, len(indexes))
	for _, idx := range indexes {
		out = append(out, &pb.IndexDataset{
			IndexCode:        idx.IndexCode,
			Market:           string(idx.Market),
			IndexName:        idx.IndexName,
			DataCompleteness: idx.DataCompleteness,
			LastSyncTime:     tsPtr(idx.LastSyncTime),
			SyncStatus:       idx.SyncStatus,
		})
	}
	return &pb.ListIndexesResponse{Indexes: out}, nil
}

func (s *Services) ListStocks(ctx context.Context, req *pb.ListStocksRequest) (*pb.ListStocksResponse, error) {
	filter := data.ListStocksFilter{
		Search: req.GetSearch(),
		Sort:   req.GetSort(),
	}
	if req.GetMarket() != "" {
		m := data.Market(req.GetMarket())
		filter.Market = &m
	}
	if req.GetStatus() != "" {
		st := data.StockSyncStatus(req.GetStatus())
		filter.Status = &st
	}
	stocks, err := s.Store.ListStocks(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list stocks: %v", err)
	}
	out := make([]*pb.StockListItem, 0, len(stocks))
	for _, st := range stocks {
		out = append(out, &pb.StockListItem{
			StockCode:    st.StockCode,
			StockName:    st.StockName,
			Market:       string(st.Market),
			SyncStatus:   string(st.SyncStatus),
			Completeness: st.Completeness,
			Open:         st.Open,
			High:         st.High,
			Low:          st.Low,
			Close:        st.Close,
			Volume:       st.Volume,
			Change:       st.Change,
			LastUpdate:   tsPtr(st.LastUpdate),
		})
	}
	return &pb.ListStocksResponse{Stocks: out}, nil
}

func (s *Services) ListDailyBars(ctx context.Context, req *pb.ListDailyBarsRequest) (*pb.ListDailyBarsResponse, error) {
	params := data.ListDailyBarsParams{
		StockCode: req.GetStockCode(),
		Limit:     int(req.GetLimit()),
	}
	if req.GetStartDate() != "" {
		t, err := data.ParseDate(req.GetStartDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		params.StartDate = &t
	}
	if req.GetEndDate() != "" {
		t, err := data.ParseDate(req.GetEndDate())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		params.EndDate = &t
	}
	bars, err := s.Store.ListDailyBars(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list daily bars: %v", err)
	}
	out := make([]*pb.DailyBar, 0, len(bars))
	for _, bar := range bars {
		out = append(out, &pb.DailyBar{
			StockCode:   bar.StockCode,
			Market:      string(bar.Market),
			TradeDate:   bar.TradeDate.Format("2006-01-02"),
			Open:        bar.Open,
			High:        bar.High,
			Low:         bar.Low,
			Close:       bar.Close,
			Volume:      bar.Volume,
			Source:      bar.Source,
			DataVersion: bar.DataVersion,
		})
	}
	return &pb.ListDailyBarsResponse{Bars: out}, nil
}

func (s *Services) SyncStock(ctx context.Context, req *pb.SyncStockRequest) (*pb.Task, error) {
	task, err := s.Store.CreateStockSyncTask(ctx, req.GetStockCode(), data.TriggerManual)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "sync stock: %v", err)
	}
	return s.toTask(task), nil
}

func (s *Services) ListDataTasks(ctx context.Context, req *pb.ListDataTasksRequest) (*pb.ListDataTasksResponse, error) {
	tasks, err := s.Store.ListDataSyncTasks(ctx, int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list data tasks: %v", err)
	}
	out := make([]*pb.Task, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, s.toTask(t))
	}
	return &pb.ListDataTasksResponse{Tasks: out}, nil
}
