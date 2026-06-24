package server

import (
	"context"
	"strings"

	"github.com/gsxhnd/guanlan/internal/data"
	"github.com/gsxhnd/guanlan/internal/task"
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
	quality, _ := s.Store.GetStockQualitySummary(ctx, req.GetStockCode())
	out := make([]*pb.DailyBar, 0, len(bars))
	for _, bar := range bars {
		out = append(out, &pb.DailyBar{
			StockCode:     bar.StockCode,
			Market:        string(bar.Market),
			TradeDate:     bar.TradeDate.Format("2006-01-02"),
			Open:          bar.Open,
			High:          bar.High,
			Low:           bar.Low,
			Close:         bar.Close,
			Volume:        bar.Volume,
			Source:        bar.Source,
			DataVersion:   bar.DataVersion,
			QualityStatus: quality.QualityStatus,
		})
	}
	return &pb.ListDailyBarsResponse{Bars: out}, nil
}

func (s *Services) ListIndexConstituents(ctx context.Context, req *pb.ListIndexConstituentsRequest) (*pb.ListIndexConstituentsResponse, error) {
	items, err := s.Store.ListIndexConstituents(ctx, req.GetIndexCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list constituents: %v", err)
	}
	out := make([]*pb.IndexConstituent, 0, len(items))
	for _, item := range items {
		c := &pb.IndexConstituent{
			IndexCode: item.IndexCode,
			StockCode: item.StockCode,
			SnapDate:  item.SnapDate.Format("2006-01-02"),
			IsActive:  item.IsActive,
		}
		if item.Weight != nil {
			c.Weight = *item.Weight
		}
		out = append(out, c)
	}
	return &pb.ListIndexConstituentsResponse{Constituents: out}, nil
}

func (s *Services) InitTrainingIndex(ctx context.Context, req *pb.InitTrainingIndexRequest) (*pb.Task, error) {
	code := req.GetIndexCode()
	if code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "index_code is required")
	}
	taskRec, err := s.Store.CreateTask(ctx, data.TaskTypeDataSync, code, data.TriggerManual, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create init task: %v", err)
	}
	if err := task.InitTrainingData(ctx, s.Python, code); err != nil {
		reason := err.Error()
		_ = s.Store.UpdateTaskStatus(ctx, taskRec.TaskID, data.TaskStatusFailed, &reason, nil)
		return nil, status.Errorf(codes.Internal, "init training: %v", err)
	}
	version := "training-" + code
	_ = s.Store.UpdateTaskStatus(ctx, taskRec.TaskID, data.TaskStatusSuccess, nil, &version)
	updated, err := s.Store.GetTask(ctx, taskRec.TaskID)
	if err != nil {
		return s.toTask(taskRec), nil
	}
	return s.toTask(updated), nil
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

func (s *Services) ListStockPool(ctx context.Context, req *pb.ListStockPoolRequest) (*pb.ListStockPoolResponse, error) {
	items, err := s.Store.ListStockPool(ctx, data.StockPoolListOptions{
		Source:        req.GetSource(),
		DailySyncOnly: req.GetDailySyncOnly(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list stock pool: %v", err)
	}
	out := make([]*pb.StockPoolItem, 0, len(items))
	for _, item := range items {
		out = append(out, toStockPoolItem(item))
	}
	return &pb.ListStockPoolResponse{Items: out, Total: int32(len(out))}, nil
}

func (s *Services) UpsertStockPoolItem(ctx context.Context, req *pb.UpsertStockPoolItemRequest) (*pb.StockPoolItem, error) {
	var entry data.StockPoolEntry
	var err error
	market := data.Market(strings.TrimSpace(req.GetMarket()))
	if yf := strings.TrimSpace(req.GetYfinanceSymbol()); yf != "" {
		entry, err = data.ParsePoolSymbol(yf, market)
	} else if original := strings.TrimSpace(req.GetOriginalCode()); original != "" {
		symbol := original
		if exchange := strings.TrimSpace(req.GetExchange()); exchange != "" {
			symbol = original + "." + exchange
		}
		entry, err = data.ParsePoolSymbol(symbol, market)
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "yfinance_symbol or original_code is required")
	}
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if original := strings.TrimSpace(req.GetOriginalCode()); original != "" {
		entry.OriginalCode = original
	}
	if name := strings.TrimSpace(req.GetStockName()); name != "" {
		entry.StockName = name
	}
	if currency := strings.TrimSpace(req.GetCurrency()); currency != "" {
		entry.Currency = currency
	}
	if exchange := strings.TrimSpace(req.GetExchange()); exchange != "" {
		entry.Exchange = exchange
	}
	entry.Source = data.PoolSourceAPIManual
	entry.IsActive = true
	entry.SyncDaily = true

	if err := s.Store.UpsertStockPoolEntry(ctx, entry); err != nil {
		return nil, status.Errorf(codes.Internal, "upsert stock pool item: %v", err)
	}
	saved, ok, err := s.Store.GetStockPoolEntry(ctx, entry.YfinanceSymbol)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get stock pool item: %v", err)
	}
	if !ok {
		return nil, status.Errorf(codes.Internal, "stock pool item not found after upsert")
	}
	return toStockPoolItem(saved), nil
}

func toStockPoolItem(item data.StockPoolEntry) *pb.StockPoolItem {
	return &pb.StockPoolItem{
		YfinanceSymbol: item.YfinanceSymbol,
		OriginalCode:   item.OriginalCode,
		Market:         string(item.Market),
		StockName:      item.StockName,
		Exchange:       item.Exchange,
		Currency:       item.Currency,
		Source:         string(item.Source),
		IsActive:       item.IsActive,
		SyncDaily:      item.SyncDaily,
		CreatedAt:      tsPtr(&item.CreatedAt),
		UpdatedAt:      tsPtr(&item.UpdatedAt),
	}
}
