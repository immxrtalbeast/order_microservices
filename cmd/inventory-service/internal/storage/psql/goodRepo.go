package psql

import (
	"context"
	"errors"

	"immxrtalbeast/order_microservices/inventory-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GoodRepository struct {
	db *gorm.DB
}

func NewGoodRepository(db *gorm.DB) *GoodRepository {
	return &GoodRepository{db: db}
}

func (r *GoodRepository) SaveGood(ctx context.Context, good *domain.Good) error {
	err := r.db.WithContext(ctx).Create(&good).Error
	return err
}

func (r *GoodRepository) DeleteGood(ctx context.Context, goodID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", goodID).Delete(&domain.Good{}).Error
}

func (r *GoodRepository) ListGoods(ctx context.Context) ([]*domain.Good, error) {
	var goods []*domain.Good
	err := r.db.WithContext(ctx).
		Model(&domain.Good{}).
		Scan(&goods).
		Error
	return goods, err
}

func (r *GoodRepository) UpdateGood(ctx context.Context, good *domain.Good) error {
	result := r.db.WithContext(ctx).Model(&domain.Good{}).
		Where("id = ?", good.ID).
		Omit("id").
		Updates(&good)

	return result.Error
}

func (r *GoodRepository) ReserveProducts(ctx context.Context, orderItems []domain.OrderItem) (int, error) {
	var total int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		goodIDs := make([]uuid.UUID, len(orderItems))
		for i, item := range orderItems {
			goodIDs[i] = item.GoodID
		}

		var goods []domain.Good
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", goodIDs).
			Find(&goods).Error; err != nil {
			return err
		}

		goodMap := make(map[uuid.UUID]domain.Good)
		for _, good := range goods {
			goodMap[good.ID] = good
		}

		total = 0
		updates := make(map[uuid.UUID]int)

		for _, item := range orderItems {
			good, exists := goodMap[item.GoodID]
			if !exists {
				return errors.New("good not found")
			}

			total += good.Price * item.Quantity

			newQuantity := good.QuantityInStock - item.Quantity
			if newQuantity < 0 {
				return errors.New("insufficient quantity")
			}
			updates[item.GoodID] = newQuantity
		}

		for goodID, quantity := range updates {
			if err := tx.Model(&domain.Good{}).
				Where("id = ?", goodID).
				Update("quantity_in_stock", quantity).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return total, nil
}
