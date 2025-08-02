package controller

import (
	"context"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	"immxrtalbeast/order_microservices/cmd/order-service/internal/lib"
	order "immxrtalbeast/order_microservices/protos/gen/go/order"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	order.UnimplementedOrderServiceServer
	orderInteractor domain.OrderInteractor
}

func Register(gRPCServer *grpc.Server, orderInteractor domain.OrderInteractor) {
	order.RegisterOrderServiceServer(gRPCServer, &serverAPI{orderInteractor: orderInteractor})
}

func (s *serverAPI) CreateOrder(ctx context.Context, in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	if in.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}
	if len(in.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one order item is required")
	}

	userID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	domainItems := make([]domain.OrderItem, len(in.Items))
	for i, item := range in.Items {
		productID, err := uuid.Parse(item.ProductId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
		}
		if item.Quantity <= 0 {
			return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
		}

		domainItems[i] = domain.OrderItem{
			ProductID: productID,
			Quantity:  int(item.Quantity),
		}
	}

	orderID, statusStr, err := s.orderInteractor.CreateOrder(ctx, userID, domainItems)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	return &order.CreateOrderResponse{
		OrderId: orderID.String(),
		Status:  lib.ConvertStatusToProto(statusStr),
	}, nil
}

func (s *serverAPI) Order(ctx context.Context, in *order.OrderRequest) (*order.OrderResponse, error) {
	if in.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	orderID, err := uuid.Parse(in.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order ID format")
	}

	domainOrder, err := s.orderInteractor.Order(ctx, orderID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get order")
	}

	return &order.OrderResponse{
		Order: lib.ConvertOrderToProto(domainOrder),
	}, nil
}

func (s *serverAPI) ListOrders(ctx context.Context, in *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	if in.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	userID, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	if in.Limit < 0 || in.Offset < 0 {
		return nil, status.Error(codes.InvalidArgument, "limit and offset must be non-negative")
	}

	orders, err := s.orderInteractor.ListOrdersByUser(ctx, userID, int(in.Limit), int(in.Offset))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list orders")
	}

	return &order.ListOrdersResponse{
		Orders: lib.ConvertOrdersToProto(orders),
	}, nil
}
