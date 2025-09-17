package auth

import (
	"context"
	"fmt"
	"net"
	"time"

	inventory "github.com/immxrtalbeast/order_protos/gen/go/inventory"

	"github.com/google/uuid"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api inventory.InventoryClient
}

func New(ctx context.Context, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			dialer := net.Dialer{}
			return dialer.DialContext(ctx, "tcp", addr)
		}),
		grpc.WithChainUnaryInterceptor(
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Client{
		api: inventory.NewInventoryClient(conn),
	}, nil
}

func (c *Client) AddGood(ctx context.Context, name, category, description, imageLink string, price, quantityInStock int, volume int32) error {
	const op = "grpc.AddGood"

	_, err := c.api.AddGood(ctx, &inventory.AddGoodRequest{
		Name:            name,
		Category:        category,
		Description:     description,
		ImageLink:       imageLink,
		Price:           float64(price),
		Volume:          volume,
		QuantityInStock: int64(quantityInStock),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Client) ListProducts(ctx context.Context) ([]*inventory.Product, error) {
	const op = "grpc.ListProducts"

	resp, err := c.api.ListProducts(ctx, &inventory.ListProductsRequest{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resp.Products, nil
}

func (c *Client) DeleteGood(ctx context.Context, goodID uuid.UUID) error {
	const op = "grpc.DeleteGood"

	_, err := c.api.DeleteGood(ctx, &inventory.DeleteGoodRequest{
		GoodId: goodID.String(),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Client) UpdateGood(ctx context.Context, goodID uuid.UUID, name, category, description, imageLink string, price, quantityInStock int) error {
	const op = "grpc.UpdateGood"

	_, err := c.api.UpdateGood(ctx, &inventory.UpdateGoodRequest{
		Id:              goodID.String(),
		Name:            name,
		Category:        category,
		Description:     description,
		ImageLink:       imageLink,
		Price:           float64(price),
		QuantityInStock: int64(quantityInStock),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
