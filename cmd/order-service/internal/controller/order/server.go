package controller

import (
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	order "immxrtalbeast/order_microservices/protos/gen/go/order"
)

type serverAPI struct {
	order.UnimplementedOrderServiceServer
	orderInteractor domain.OrderInteractor
}
