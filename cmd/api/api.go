package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gsxhnd/guanlan/internal/data"
	"github.com/gsxhnd/guanlan/internal/server"
	"github.com/gsxhnd/guanlan/internal/task"
	pb "github.com/gsxhnd/guanlan/internal/proto/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type runtimeState struct {
	cancelScheduled context.CancelFunc
}

func NewStore(cfg Config) (*data.Store, error) {
	return data.Open(cfg.DBPath)
}

func NewServices(store *data.Store) *server.Services {
	return &server.Services{Store: store}
}

func NewScheduler(store *data.Store, cfg Config) *task.Scheduler {
	return task.NewScheduler(store, &task.DataSyncExecutor{Store: store}, cfg.TaskPollInterval)
}

func NewGRPCServer(svc *server.Services) *grpc.Server {
	srv := grpc.NewServer()
	pb.RegisterTaskServiceServer(srv, svc)
	pb.RegisterWatchlistServiceServer(srv, svc)
	pb.RegisterPortfolioServiceServer(srv, svc)
	pb.RegisterDataServiceServer(srv, svc)
	return srv
}

func NewListener(cfg Config) (net.Listener, error) {
	return net.Listen("tcp", cfg.GRPCAddr)
}

func Run(
	lc fx.Lifecycle,
	cfg Config,
	lis net.Listener,
	srv *grpc.Server,
	store *data.Store,
	sched *task.Scheduler,
) {
	ctx, cancel := context.WithCancel(context.Background())
	state := &runtimeState{}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Printf("api server listening on %s (db: %s)", cfg.GRPCAddr, cfg.DBPath)
			sched.Start(ctx)
			schedCtx, schedCancel := context.WithCancel(ctx)
			state.cancelScheduled = schedCancel
			go func() {
				ticker := time.NewTicker(cfg.ScheduledSyncInterval)
				defer ticker.Stop()
				for {
					select {
					case <-schedCtx.Done():
						return
					case <-ticker.C:
						if err := task.ScheduledSync(schedCtx, store); err != nil {
							log.Printf("scheduled sync: %v", err)
						}
					}
				}
			}()
			go func() {
				if err := srv.Serve(lis); err != nil {
					log.Printf("serve: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			if state.cancelScheduled != nil {
				state.cancelScheduled()
			}
			sched.Stop()
			stopped := make(chan struct{})
			go func() {
				srv.GracefulStop()
				close(stopped)
			}()
			select {
			case <-stopped:
			case <-time.After(5 * time.Second):
				srv.Stop()
			}
			return store.Close()
		},
	})
}
