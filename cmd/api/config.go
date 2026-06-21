package main

import "flag"

type Config struct {
	GRPCAddr string
}

func NewConfig() Config {
	grpcAddr := flag.String("addr", ":50051", "gRPC listen address")
	flag.Parse()

	return Config{
		GRPCAddr: *grpcAddr,
	}
}
