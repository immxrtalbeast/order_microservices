package order

import (
	"context"
	"fmt"
	order "immxrtalbeast/order_microservices/protos/gen/go/order"
	"net"
	"time"

	"github.com/google/uuid"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api order.OrderServiceClient
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
		api: order.NewOrderServiceClient(conn),
	}, nil

}

func (c *Client) CreateOrder(ctx context.Context, userID string, items []*order.OrderItem) (*order.CreateOrderResponse, error) {
	const op = "grpc.CreateOrder"

	resp, err := c.api.CreateOrder(ctx, &order.CreateOrderRequest{
		UserId: userID,
		Items:  items,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}

func (c *Client) DeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	const op = "grpc.DeleteOrder"

	_, err := c.api.DeleteOrder(ctx, &order.DeleteOrderRequest{
		OrderId: orderID.String(),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*order.OrderResponse, error) {
	const op = "grpc.GetOrder"

	resp, err := c.api.Order(ctx, &order.OrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}

func (c *Client) ListOrdersByUser(ctx context.Context, userID string, limit, offset int32) (*order.ListOrdersResponse, error) {
	const op = "grpc.ListOrdersByUser"

	resp, err := c.api.ListOrders(ctx, &order.ListOrdersRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}
