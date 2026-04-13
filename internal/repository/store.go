package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"kanalegeleri/go-app/internal/domain"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// AutoMigrate creates or updates tables based on domain model structs.
func (s *Store) AutoMigrate() error {
	return s.db.AutoMigrate(
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
	)
}

// --- Products ---

func (s *Store) AllProducts(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	err := s.db.WithContext(ctx).Order("id asc").Find(&products).Error
	return products, err
}

func (s *Store) GetProductByID(ctx context.Context, id int) (domain.Product, bool, error) {
	var product domain.Product
	err := s.db.WithContext(ctx).First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Product{}, false, nil
	}
	if err != nil {
		return domain.Product{}, false, err
	}
	return product, true, nil
}

func (s *Store) CreateProduct(ctx context.Context, product domain.Product) error {
	return s.db.WithContext(ctx).Create(&product).Error
}

func (s *Store) UpdateProduct(ctx context.Context, product domain.Product) error {
	result := s.db.WithContext(ctx).
		Model(&domain.Product{}).
		Where("id = ?", product.ID).
		Updates(map[string]any{
			"name":        product.Name,
			"description": product.Description,
			"category":    product.Category,
			"image_url":   product.ImageURL,
			"price":       product.Price,
			"updated_at":  time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteProduct(ctx context.Context, id int) error {
	return s.db.WithContext(ctx).Delete(&domain.Product{}, id).Error
}

// --- Orders ---

func (s *Store) CreateOrder(ctx context.Context, input domain.CreateOrderRequest) (domain.Order, error) {
	order := domain.Order{
		FormType:        input.FormType,
		CustomerName:    input.CustomerName,
		CustomerEmail:   input.CustomerEmail,
		CustomerPhone:   input.CustomerPhone,
		CustomerAddress: input.CustomerAddress,
		FullAddress:     input.FullAddress,
		Note:            input.Note,
		Items:           input.Items,
	}
	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

func (s *Store) AllOrders(ctx context.Context) ([]domain.Order, error) {
	var orders []domain.Order
	err := s.db.WithContext(ctx).
		Preload("Items").
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}
