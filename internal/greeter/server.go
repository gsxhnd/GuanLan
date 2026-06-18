package greeter

import (
	"context"
	"fmt"

	gen "github.com/gsxhnd/guanlan/gen"
)

type Server struct {
	gen.UnimplementedGreeterServer
}

func (s *Server) SayHello(ctx context.Context, req *gen.HelloRequest) (*gen.HelloReply, error) {
	name := req.GetName()
	if name == "" {
		name = "world"
	}
	return &gen.HelloReply{
		Message: fmt.Sprintf("Hello, %s!", name),
	}, nil
}
