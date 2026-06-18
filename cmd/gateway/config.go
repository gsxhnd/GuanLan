package main

import "flag"

type Config struct {
	HTTPAddr string
	GRPCAddr string
}

func NewConfig() Config {
	httpAddr := flag.String("http", ":8080", "HTTP listen address")
	grpcAddr := flag.String("grpc", "localhost:50051", "gRPC backend address")
	flag.Parse()

	return Config{
		HTTPAddr: *httpAddr,
		GRPCAddr: *grpcAddr,
	}
}
