package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/gsxhnd/guanlan/internal/proto/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServeMux(cfg Config) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	ctx := context.Background()
	registrars := []func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error{
		pb.RegisterTaskServiceHandlerFromEndpoint,
		pb.RegisterWatchlistServiceHandlerFromEndpoint,
		pb.RegisterPortfolioServiceHandlerFromEndpoint,
		pb.RegisterDataServiceHandlerFromEndpoint,
	}
	for _, register := range registrars {
		if err := register(ctx, mux, cfg.GRPCAddr, opts); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

func NewApp(mux *runtime.ServeMux) *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	app.All("/api/*", adaptor.HTTPHandler(mux))
	return app
}

func Run(lc fx.Lifecycle, cfg Config, app *fiber.App) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("gRPC gateway listening on %s (backend %s)", cfg.HTTPAddr, cfg.GRPCAddr)
			go func() {
				if err := app.Listen(cfg.HTTPAddr); err != nil {
					log.Printf("serve: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.ShutdownWithContext(ctx)
		},
	})
}
