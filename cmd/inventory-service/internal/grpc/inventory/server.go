package grpc

import (
	"context"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/lib"
	inventory "immxrtalbeast/order_microservices/protos/gen/go/inventory"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	inventory.UnimplementedInventoryServer
	inventoryController InventoryController
}

type InventoryController interface {
	AddGood(ctx context.Context, name string, description string, imageLink string, price int, quantityInStock int) error
	ListProducts(ctx context.Context) ([]*domain.Good, error)
	DeleteGood(ctx context.Context, goodID uuid.UUID) error
	UpdateGood(ctx context.Context, goodID uuid.UUID, name string, description string, imageLink string, price int, quantityInStock int) error
}

func Register(gRPCServer *grpc.Server, inventoryController InventoryController) {
	inventory.RegisterInventoryServer(gRPCServer, &serverAPI{inventoryController: inventoryController})
}

func (s *serverAPI) AddGood(ctx context.Context, in *inventory.AddGoodRequest) (*inventory.AddGoodResponse, error) {
	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if in.Price == 0 || in.Price < 0 {
		return nil, status.Error(codes.InvalidArgument, "price should be greater than 0")
	}
	if in.QuantityInStock < 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity should be equal/greater than 0")
	}

	if err := s.inventoryController.AddGood(ctx, in.Name, in.Description, in.ImageLink, int(in.Price), int(in.QuantityInStock)); err != nil {
		return nil, status.Error(codes.Internal, "failed to save good")
	}
	return &inventory.AddGoodResponse{Success: true}, nil
}

func (s *serverAPI) ListProducts(ctx context.Context, in *inventory.ListProductsRequest) (*inventory.ListProductsResponse, error) {
	goods, err := s.inventoryController.ListProducts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get goods")
	}
	products := lib.ConvertGoodToProduct(goods)

	return &inventory.ListProductsResponse{Products: products}, nil
}

func (s *serverAPI) DeleteGood(ctx context.Context, in *inventory.DeleteGoodRequest) (*inventory.DeleteGoodResponse, error) {
	good_id, _ := uuid.Parse(in.GoodId)
	if err := s.inventoryController.DeleteGood(ctx, good_id); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete good")
	}
	return &inventory.DeleteGoodResponse{Success: true}, nil
}

func (s *serverAPI) UpdateGood(ctx context.Context, in *inventory.UpdateGoodRequest) (*inventory.UpdateGoodResponse, error) {
	good_id, _ := uuid.Parse(in.Id)
	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if in.Price == 0 || in.Price < 0 {
		return nil, status.Error(codes.InvalidArgument, "price should be greater than 0")
	}
	if in.QuantityInStock < 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity should be equal/greater than 0")
	}
	if err := s.inventoryController.UpdateGood(ctx, good_id, in.Name, in.Description, in.ImageLink, int(in.Price), int(in.QuantityInStock)); err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to update good")
	}
	return &inventory.UpdateGoodResponse{Success: true}, nil
}
