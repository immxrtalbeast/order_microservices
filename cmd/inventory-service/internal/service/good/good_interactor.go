package good

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"

	"github.com/google/uuid"
)

type GoodInteractor struct {
	goodRepo domain.GoodRepository
}

func NewGoodInteractor(goodRepo domain.GoodRepository) *GoodInteractor {
	return &GoodInteractor{goodRepo: goodRepo}
}

func (gi *GoodInteractor) AddGood(ctx context.Context, name string, description string, imageLink string, price int, quantityInStock int) error {
	const op = "service.good.save"
	good := &domain.Good{
		Name:            name,
		Description:     description,
		ImageLink:       imageLink,
		Price:           price,
		QuantityInStock: quantityInStock,
	}

	if err := gi.goodRepo.SaveGood(ctx, good); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (gi *GoodInteractor) ListProductssss(ctx context.Context) ([]*domain.Good, error) {
	const op = "service.good.list"
	goods, err := gi.goodRepo.ListGoods(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return goods, nil
}

func (gi *GoodInteractor) DeleteGood(ctx context.Context, goodID uuid.UUID) error {
	const op = "service.good.delete"
	if err := gi.goodRepo.DeleteGood(ctx, goodID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (gi *GoodInteractor) UpdateGood(ctx context.Context, goodID uuid.UUID, name string, description string, imageLink string, price int, quantityInStock int) error {
	const op = "service.good.update"
	good := &domain.Good{
		ID:              goodID,
		Name:            name,
		Description:     description,
		ImageLink:       imageLink,
		Price:           price,
		QuantityInStock: quantityInStock,
	}
	if err := gi.goodRepo.UpdateGood(ctx, good); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
