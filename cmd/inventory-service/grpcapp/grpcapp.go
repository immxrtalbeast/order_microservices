package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	inventorygrpc "immxrtalbeast/order_microservices/inventory-service/internal/grpc/inventory"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcApp struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int // Порт, на котором будет работать grpc-сервер
}

func New(log *slog.Logger, inventoryInteractor domain.InventoryInteractor, port int) *GrpcApp {

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()))

	inventorygrpc.Register(gRPCServer, inventoryInteractor)

	return &GrpcApp{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *GrpcApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *GrpcApp) Run() error {
	const op = "grpcapp.Run"
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
