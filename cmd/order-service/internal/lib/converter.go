package lib

import (
	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"
	order "immxrtalbeast/order_microservices/protos/gen/go/order"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertOrderToProto(o domain.Order) *order.Order {
	items := make([]*order.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &order.OrderItem{
			ProductId: item.ProductID.String(),
			Quantity:  int32(item.Quantity),
		}
	}

	return &order.Order{
		Id:        o.ID.String(),
		UserId:    o.UserID.String(),
		Items:     items,
		Total:     float32(o.Total),
		Status:    ConvertStatusToProto(o.Status),
		CreatedAt: timestamppb.New(o.CreatedAt),
		UpdatedAt: timestamppb.New(o.UpdatedAt),
	}
}

func ConvertOrdersToProto(orders []domain.Order) []*order.Order {
	pbOrders := make([]*order.Order, len(orders))
	for i, o := range orders {
		pbOrders[i] = ConvertOrderToProto(o)
	}
	return pbOrders
}

func ConvertStatusToProto(status string) order.OrderStatus {
	switch status {
	case "PROCESSING":
		return order.OrderStatus_PROCESSING
	case "COMPLETED":
		return order.OrderStatus_COMPLETED
	case "CANCELLED":
		return order.OrderStatus_CANCELLED
	default:
		return order.OrderStatus_CREATED
	}
}

func ConvertItemstoEventItems(items []domain.OrderItem) []domain.OrderItemEvent {
	order_items := make([]domain.OrderItemEvent, len(items))
	for i, item := range items {
		order_items[i] = domain.OrderItemEvent{
			GoodID:   item.ProductID,
			Quantity: item.Quantity,
		}
	}

	return order_items
}
