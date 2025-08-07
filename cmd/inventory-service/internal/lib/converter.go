package lib

import (
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	inventory "immxrtalbeast/order_microservices/protos/gen/go/inventory"
)

func ConvertGoodToProduct(dbGoods []*domain.Good) []*inventory.Product {
	pbProducts := make([]*inventory.Product, 0, len(dbGoods))
	for _, g := range dbGoods {
		pbProduct := &inventory.Product{
			Id:              g.ID.String(),
			Name:            g.Name,
			ImageLink:       g.ImageLink,
			Description:     g.Description,
			Price:           float64(g.Price),
			Volume:          int32(g.Volume),
			QuantityInStock: int64(g.QuantityInStock),
		}
		pbProducts = append(pbProducts, pbProduct)
	}
	return pbProducts
}
