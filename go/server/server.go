package main

import (
	"context"

	"github.com/olte36/grpc-monorepo-example/genproto/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	api.UnimplementedStockServiceServer
}

func (*Server) List(context.Context, *emptypb.Empty) (*api.ListResponse, error) {

}
