package main

import (
	"context"
	"log"
	"net"

	gen "github.com/gsxhnd/guanlan/gen/v1"
	"github.com/gsxhnd/guanlan/internal/greeter"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func NewListener(cfg Config) (net.Listener, error) {
	return net.Listen("tcp", cfg.GRPCAddr)
}

func NewGRPCServer() *grpc.Server {
	srv := grpc.NewServer()
	gen.RegisterGreeterServer(srv, &greeter.Server{})
	return srv
}

func Run(lc fx.Lifecycle, cfg Config, lis net.Listener, srv *grpc.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("api server listening on %s", cfg.GRPCAddr)
			go func() {
				if err := srv.Serve(lis); err != nil {
					log.Printf("serve: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopped := make(chan struct{})
			go func() {
				srv.GracefulStop()
				close(stopped)
			}()
			select {
			case <-stopped:
				return nil
			case <-ctx.Done():
				srv.Stop()
				return ctx.Err()
			}
		},
	})
}
