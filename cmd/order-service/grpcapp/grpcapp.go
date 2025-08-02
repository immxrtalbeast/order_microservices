package grpcapp

import (
	"log/slog"

	"google.golang.org/grpc"
)

type GrpcApp struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}
