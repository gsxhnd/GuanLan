package httpx

import (
	"net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

// NewServer builds the unified HTTP server (CORS, healthz, metrics).
// Service routes are registered by callers via pb.Register*HTTPServer.
func NewServer(addr string, opts ...khttp.ServerOption) *khttp.Server {
	base := []khttp.ServerOption{
		khttp.Address(addr),
		khttp.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{
				"http://localhost:1420",
				"http://127.0.0.1:1420",
				"http://localhost:5173",
				"http://127.0.0.1:5173",
			}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		)),
	}
	base = append(base, opts...)
	srv := khttp.NewServer(base...)

	r := srv.Route("/")
	r.GET("/healthz", func(ctx khttp.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})
	r.GET("/metrics", func(ctx khttp.Context) error {
		// Placeholder until OTEL/Prometheus scrape is wired; keeps the route live.
		return ctx.String(http.StatusOK, "# guanlan metrics\n")
	})

	return srv
}
