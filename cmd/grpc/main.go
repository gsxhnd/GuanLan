package main

import (
	"flag"
	"log"
	"net"

	"github.com/gsxhnd/guanlan/internal/greeter"
	gen "github.com/gsxhnd/guanlan/gen"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":50051", "gRPC listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	gen.RegisterGreeterServer(srv, &greeter.Server{})

	log.Printf("gRPC server listening on %s", *addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
