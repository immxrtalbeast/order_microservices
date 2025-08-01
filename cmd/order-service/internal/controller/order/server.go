package controller

import (
	order "immxrtalbeast/order_microservices/protos/gen/go/order"
)

type serverAPI struct {
	order.UnimplementedOrderServer
	order Order
}
