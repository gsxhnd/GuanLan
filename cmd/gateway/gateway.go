package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gen "github.com/gsxhnd/guanlan/gen/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type fiberLocalKey string

const demoLocalKey = "123"

func bridgeLocalToUserContext(c fiber.Ctx, key string) {
	if v := c.Locals(key); v != nil {
		c.SetContext(context.WithValue(c.Context(), fiberLocalKey(key), v))
	}
}

func fiberLocalsToMetadata(_ context.Context, req *http.Request) metadata.MD {
	fctx, ok := adaptor.LocalContextFromHTTPRequest(req)
	if !ok {
		return nil
	}
	if v := fctx.Value(fiberLocalKey(demoLocalKey)); v != nil {
		return metadata.Pairs(demoLocalKey, fmt.Sprint(v))
	}
	return nil
}

func NewServeMux(cfg Config) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux(runtime.WithMetadata(fiberLocalsToMetadata))

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := gen.RegisterGreeterHandlerFromEndpoint(context.Background(), mux, cfg.GRPCAddr, opts); err != nil {
		return nil, err
	}

	return mux, nil
}

func NewApp(mux *runtime.ServeMux) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		var a = c.Get("xxx", "0")
		c.Locals(demoLocalKey, a)
		fmt.Println("demoLocalKey", a)
		bridgeLocalToUserContext(c, demoLocalKey)
		return c.Next()
	})
	app.All("/v1/*", adaptor.HTTPHandlerWithContext(mux))
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
