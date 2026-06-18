package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	gen "github.com/gsxhnd/guanlan/gen"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	httpAddr := flag.String("http", ":8080", "HTTP listen address")
	grpcAddr := flag.String("grpc", "localhost:50051", "gRPC backend address")
	flag.Parse()

	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := gen.RegisterGreeterHandlerFromEndpoint(ctx, mux, *grpcAddr, opts); err != nil {
		log.Fatalf("register gateway handlers: %v", err)
	}

	log.Printf("gRPC gateway listening on %s (backend %s)", *httpAddr, *grpcAddr)
	if err := http.ListenAndServe(*httpAddr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
