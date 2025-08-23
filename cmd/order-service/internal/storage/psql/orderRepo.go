package psql

import (
	"context"
	"errors"
	"fmt"

	"immxrtalbeast/order_microservices/cmd/order-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) SaveOrder(ctx context.Context, order *domain.Order) (uuid.UUID, error) {
	if len(order.Items) == 0 {
		return uuid.Nil, errors.New("order items cannot be empty")
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return uuid.Nil, tx.Error
	}

	// 1. Сохраняем основной заказ
	if err := tx.Omit("Items").Create(order).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	// 2. Подготовка элементов
	for i := range order.Items {
		order.Items[i].OrderID = order.ID // Устанавливаем связь
		order.Items[i].ID = uuid.Nil      // Сбрасываем ID для генерации нового

	}

	// 3. Массовое сохранение элементов
	if err := tx.Create(&order.Items).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	// 4. Фиксация транзакции
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	return order.ID, nil
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", orderID).Delete(&domain.Order{}).Error
}

func (r *OrderRepository) GetOrder(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	var order domain.Order

	result := r.db.WithContext(ctx).Preload("Items").Where("id = ?", orderID).First(&order)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.Order{}, fmt.Errorf("order not found")
		}
		return domain.Order{}, fmt.Errorf("database error: %w", result.Error)
	}

	return order, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", orderID).Update("status", status)
	return result.Error
}

func (r *OrderRepository) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(orders).Error
	return orders, err
}

func (r *OrderRepository) SetTotalSum(ctx context.Context, orderID uuid.UUID, sum int) error {
	result := r.db.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", orderID).Update("total", sum)
	return result.Error
}
