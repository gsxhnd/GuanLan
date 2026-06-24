package server

import pb "github.com/gsxhnd/guanlan/internal/proto/v1"

var (
	_ pb.TaskServiceServer      = (*Services)(nil)
	_ pb.WatchlistServiceServer = (*Services)(nil)
	_ pb.PortfolioServiceServer = (*Services)(nil)
	_ pb.DataServiceServer      = (*Services)(nil)
)
