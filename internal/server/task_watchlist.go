package server

import (
	"context"

	"github.com/gsxhnd/guanlan/internal/data"
	pb "github.com/gsxhnd/guanlan/internal/proto/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Services struct {
	Store *data.Store
}

func (s *Services) toTask(t data.DataSyncTask) *pb.Task {
	return &pb.Task{
		TaskId:        t.TaskID,
		TaskType:      string(t.TaskType),
		TargetObject:  t.TargetObject,
		TriggerMethod: string(t.TriggerMethod),
		Status:        string(t.Status),
		CreatedAt:     tsPtr(&t.CreatedAt),
		StartedAt:     tsPtr(t.StartedAt),
		EndedAt:       tsPtr(t.EndedAt),
		RetryCount:    int32(t.RetryCount),
		FailureReason: strPtr(t.FailureReason),
		LogRef:        strPtr(t.LogRef),
		DataVersion:   strPtr(t.DataVersion),
	}
}

// --- TaskService ---

func (s *Services) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	var taskType *data.TaskType
	if req.GetTaskType() != "" {
		t := data.TaskType(req.GetTaskType())
		taskType = &t
	}
	tasks, err := s.Store.ListTasks(ctx, taskType, int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list tasks: %v", err)
	}
	out := make([]*pb.Task, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, s.toTask(t))
	}
	return &pb.ListTasksResponse{Tasks: out}, nil
}

func (s *Services) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	task, err := s.Store.GetTask(ctx, req.GetTaskId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "get task: %v", err)
	}
	return s.toTask(task), nil
}

func (s *Services) RetryTask(ctx context.Context, req *pb.RetryTaskRequest) (*pb.RetryTaskResponse, error) {
	task, err := s.Store.RetryTask(ctx, req.GetTaskId())
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "retry task: %v", err)
	}
	return &pb.RetryTaskResponse{Task: s.toTask(task)}, nil
}

// --- WatchlistService ---

func (s *Services) toWatchlistItem(w data.WatchlistItem) *pb.WatchlistItem {
	return &pb.WatchlistItem{
		StockCode:    w.StockCode,
		Market:       string(w.Market),
		Tags:         w.Tags,
		Priority:     int32(w.Priority),
		Notes:        w.Notes,
		IsActive:     w.IsActive,
		AddedAt:      tsPtr(&w.AddedAt),
		RemovedAt:    tsPtr(w.RemovedAt),
		Source:       w.Source,
		LastAction:   w.LastAction,
		LastActionAt: tsPtr(w.LastActionAt),
		SyncStatus:   w.SyncStatus,
		Completeness: w.Completeness,
	}
}

func (s *Services) ListWatchlist(ctx context.Context, req *pb.ListWatchlistRequest) (*pb.ListWatchlistResponse, error) {
	items, err := s.Store.ListWatchlistItems(ctx, req.GetActiveOnly())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list watchlist: %v", err)
	}
	out := make([]*pb.WatchlistItem, 0, len(items))
	for _, item := range items {
		out = append(out, s.toWatchlistItem(item))
	}
	return &pb.ListWatchlistResponse{Items: out}, nil
}

func (s *Services) AddWatchlistItem(ctx context.Context, req *pb.AddWatchlistItemRequest) (*pb.WatchlistItem, error) {
	item, err := s.Store.AddWatchlistItem(ctx, data.WatchlistItem{
		StockCode: req.GetStockCode(),
		Market:    data.Market(req.GetMarket()),
		Tags:      req.GetTags(),
		Priority:  int(req.GetPriority()),
		Notes:     req.GetNotes(),
		IsActive:  req.GetIsActive(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "add watchlist item: %v", err)
	}

	// 若 DuckDB 无日频数据则自动创建获取任务
	statusRow, _ := s.Store.ListStocks(ctx, data.ListStocksFilter{Search: item.StockCode})
	needsSync := true
	for _, st := range statusRow {
		if st.StockCode == item.StockCode && st.SyncStatus == data.StockStatusReady {
			needsSync = false
			break
		}
	}
	if needsSync {
		_, _ = s.Store.CreateStockSyncTask(ctx, item.StockCode, data.TriggerManual)
	}

	return s.toWatchlistItem(item), nil
}

func (s *Services) RemoveWatchlistItem(ctx context.Context, req *pb.RemoveWatchlistItemRequest) (*pb.WatchlistItem, error) {
	item, err := s.Store.RemoveWatchlistItem(ctx, req.GetStockCode(), req.GetHardDelete())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "remove watchlist item: %v", err)
	}
	return s.toWatchlistItem(item), nil
}
