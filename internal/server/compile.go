package server

import pb "github.com/gsxhnd/guanlan/internal/proto/quant/v1"

var (
	_ pb.TaskServiceServer      = (*Services)(nil)
	_ pb.WatchlistServiceServer = (*Services)(nil)
	_ pb.PortfolioServiceServer = (*Services)(nil)
	_ pb.DataServiceServer      = (*Services)(nil)
	_ pb.AnalysisServiceServer  = (*Services)(nil)

	_ pb.TaskServiceHTTPServer      = (*Services)(nil)
	_ pb.WatchlistServiceHTTPServer = (*Services)(nil)
	_ pb.PortfolioServiceHTTPServer = (*Services)(nil)
	_ pb.DataServiceHTTPServer      = (*Services)(nil)
	_ pb.AnalysisServiceHTTPServer  = (*Services)(nil)
)
