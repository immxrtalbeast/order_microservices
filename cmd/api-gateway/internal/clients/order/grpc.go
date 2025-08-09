package order

import (
	"context"
	order "immxrtalbeast/order_microservices/protos/gen/go/order"
)

type Client struct {
	api order.OrderServiceClient
}

func New(ctx context.Context)
