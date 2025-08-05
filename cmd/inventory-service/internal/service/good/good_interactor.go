package good

import (
	"context"
	"fmt"
	"immxrtalbeast/order_microservices/inventory-service/internal/domain"
	"immxrtalbeast/order_microservices/inventory-service/internal/lib/logger/sl"
	"log/slog"

	"github.com/google/uuid"
)

type GoodInteractor struct {
	log      *slog.Logger
	goodRepo domain.GoodRepository
}

func NewGoodInteractor(goodRepo domain.GoodRepository, log *slog.Logger) *GoodInteractor {
	return &GoodInteractor{goodRepo: goodRepo, log: log}
}

func (gi *GoodInteractor) AddGood(ctx context.Context, name string, description string, imageLink string, price int, quantityInStock int) error {
	const op = "service.good.save"
	log := gi.log.With(
		slog.String("op", op),
		slog.String("good", name),
	)

	log.Info("adding good")
	good := &domain.Good{
		Name:            name,
		Description:     description,
		ImageLink:       imageLink,
		Price:           price,
		QuantityInStock: quantityInStock,
	}

	if err := gi.goodRepo.SaveGood(ctx, good); err != nil {
		log.Error("failed to save good", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Good saved")
	return nil
}

func (gi *GoodInteractor) ListProducts(ctx context.Context) ([]*domain.Good, error) {
	const op = "service.good.list"
	log := gi.log.With(
		slog.String("op", op),
	)
	log.Info("getting list of goods")
	goods, err := gi.goodRepo.ListGoods(ctx)
	if err != nil {
		log.Error("failed to get list of goods", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("list provided")
	return goods, nil
}

func (gi *GoodInteractor) DeleteGood(ctx context.Context, goodID uuid.UUID) error {
	const op = "service.good.delete"
	log := gi.log.With(
		slog.String("op", op),
		slog.String("goodID", goodID.String()),
	)
	log.Info("deleting good")
	if err := gi.goodRepo.DeleteGood(ctx, goodID); err != nil {
		log.Error("failed to delete good", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("good deleted")
	return nil
}

func (gi *GoodInteractor) UpdateGood(ctx context.Context, goodID uuid.UUID, name string, description string, imageLink string, price int, quantityInStock int) error {
	const op = "service.good.update"
	log := gi.log.With(
		slog.String("op", op),
		slog.String("goodID", goodID.String()),
		slog.String("name", name),
		slog.String("description", description),
		slog.String("imageLink", imageLink),
		slog.Int("price", price),
		slog.Int("quantity", quantityInStock),
	)
	log.Info("updating good")
	good := &domain.Good{
		ID:              goodID,
		Name:            name,
		Description:     description,
		ImageLink:       imageLink,
		Price:           price,
		QuantityInStock: quantityInStock,
	}
	if err := gi.goodRepo.UpdateGood(ctx, good); err != nil {
		log.Error("failed to update good", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("good updated")
	return nil
}
