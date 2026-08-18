package main

import (
	"context"
	"log"

	"github.com/gsxhnd/guanlan/internal/biz"
	glcron "github.com/gsxhnd/guanlan/internal/cron"
	"github.com/gsxhnd/guanlan/internal/data"
	"github.com/gsxhnd/guanlan/internal/orchestrator"
	pb "github.com/gsxhnd/guanlan/internal/proto/quant/v1"
	"github.com/gsxhnd/guanlan/internal/server"
	httpx "github.com/gsxhnd/guanlan/internal/transport/http"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type pythonConns struct {
	crawler *grpc.ClientConn
	predict *grpc.ClientConn
}

func NewStore(cfg Config) (*data.Store, error) {
	return data.Open(cfg.DBPath)
}

func NewBiz(store *data.Store) *biz.Services {
	return biz.New(store)
}

func NewPythonClients(cfg Config) (pb.CrawlerServiceClient, pb.PredictionServiceClient, *pythonConns, error) {
	crawlerConn, err := grpc.NewClient(cfg.CrawlerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, err
	}
	predictConn, err := grpc.NewClient(cfg.PredictionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = crawlerConn.Close()
		return nil, nil, nil, err
	}
	return pb.NewCrawlerServiceClient(crawlerConn),
		pb.NewPredictionServiceClient(predictConn),
		&pythonConns{crawler: crawlerConn, predict: predictConn},
		nil
}

func NewPredictExecutor(store *data.Store, predict pb.PredictionServiceClient) *orchestrator.PredictExecutor {
	return &orchestrator.PredictExecutor{Store: store, Client: predict}
}

func NewServices(store *data.Store, bizSvc *biz.Services, predict *orchestrator.PredictExecutor) *server.Services {
	return &server.Services{Store: store, Biz: bizSvc, Predict: predict}
}

func NewScheduler(
	store *data.Store,
	crawler pb.CrawlerServiceClient,
	predict *orchestrator.PredictExecutor,
	cfg Config,
) *orchestrator.Scheduler {
	exec := &orchestrator.Dispatcher{
		Sync: &orchestrator.SyncExecutor{
			Store:        store,
			Crawler:      crawler,
			LookbackDays: cfg.SyncLookbackDays,
		},
		Predict: predict,
	}
	return orchestrator.NewScheduler(store, exec, cfg.TaskPollInterval)
}

func NewCron(bizSvc *biz.Services, cfg Config) (*glcron.Scheduler, error) {
	return glcron.New(bizSvc, cfg.DailySyncCron)
}

func NewHTTPServer(cfg Config, svc *server.Services) *khttp.Server {
	srv := httpx.NewServer(cfg.HTTPAddr)
	pb.RegisterTaskServiceHTTPServer(srv, svc)
	pb.RegisterWatchlistServiceHTTPServer(srv, svc)
	pb.RegisterPortfolioServiceHTTPServer(srv, svc)
	pb.RegisterDataServiceHTTPServer(srv, svc)
	pb.RegisterAnalysisServiceHTTPServer(srv, svc)
	return srv
}

func Run(
	lc fx.Lifecycle,
	cfg Config,
	httpSrv *khttp.Server,
	store *data.Store,
	sched *orchestrator.Scheduler,
	cronSched *glcron.Scheduler,
	conns *pythonConns,
) {
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Printf("api server listening on %s (db: %s, crawler: %s, prediction: %s)",
				cfg.HTTPAddr, cfg.DBPath, cfg.CrawlerAddr, cfg.PredictionAddr)
			sched.Start(ctx)
			cronSched.Start()
			go func() {
				if err := httpSrv.Start(ctx); err != nil {
					log.Printf("http serve: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			sched.Stop()
			cronStop := cronSched.Stop()
			select {
			case <-cronStop.Done():
			case <-stopCtx.Done():
			}
			if err := httpSrv.Stop(stopCtx); err != nil {
				log.Printf("http stop: %v", err)
			}
			if conns != nil {
				if conns.crawler != nil {
					_ = conns.crawler.Close()
				}
				if conns.predict != nil {
					_ = conns.predict.Close()
				}
			}
			return store.Close()
		},
	})
}
