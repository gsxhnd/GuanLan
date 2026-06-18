package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	gen "github.com/gsxhnd/guanlan/gen"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServeMux(cfg Config) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := gen.RegisterGreeterHandlerFromEndpoint(context.Background(), mux, cfg.GRPCAddr, opts); err != nil {
		return nil, err
	}

	return mux, nil
}

func Run(lc fx.Lifecycle, cfg Config, mux *runtime.ServeMux) {
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("gRPC gateway listening on %s (backend %s)", cfg.HTTPAddr, cfg.GRPCAddr)
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("serve: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
