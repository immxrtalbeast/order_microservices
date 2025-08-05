package grpcapp

import (
	"fmt"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	ordergrpc "immxrtalbeast/order_microservices/cmd/order-service/internal/grpc/order"
	"log/slog"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcApp struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, orderInteractor domain.OrderInteractor, port int) *GrpcApp {

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
	))

	ordergrpc.Register(gRPCServer, orderInteractor)

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

	// Создаём listener, который будет слушить TCP-сообщения, адресованные
	// Нашему gRPC-серверу
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	// Запускаем обработчик gRPC-сообщений
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
