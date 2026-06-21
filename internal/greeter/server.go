package greeter

import (
	"context"
	"fmt"

	gen "github.com/gsxhnd/guanlan/gen/v1"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	gen.UnimplementedGreeterServer
}

func (s *Server) SayHello(ctx context.Context, req *gen.HelloRequest) (*gen.HelloReply, error) {
	name := req.GetName()
	if name == "" {
		name = "world"
	}
	fmt.Println("SayHello", name)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Println("SayHello metadata value", md.Get("123"))
	}
	return &gen.HelloReply{
		Message: fmt.Sprintf("Hello, %s!", name),
	}, nil
}
