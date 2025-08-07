package psql

import (
	"context"

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
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return uuid.Nil, tx.Error
	}

	// 1. Сохраняем основной заказ
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	// 2. Подготовка элементов
	for i := range order.Items {
		order.Items[i].OrderID = order.ID // Устанавливаем связь

		// ⭐ ВАЖНО: Сбрасываем ID чтобы БД сгенерировала новый UUID ⭐
		order.Items[i].ID = uuid.Nil
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

func (r *OrderRepository) GetOrder(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	var order domain.Order

	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(order).Error
	return order, err
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", orderID).Update("status", status)
	return result.Error
}

func (r *OrderRepository) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Scan(orders).Error
	return orders, err
}
