package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"golang.org/x/crypto/bcrypt"
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
		&domain.Category{},
		&domain.Admin{},
	)
}

// --- Admin ---

func (s *Store) GetAdminByUsername(ctx context.Context, username string) (domain.Admin, bool, error) {
	var admin domain.Admin
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Admin{}, false, nil
	}
	return admin, err == nil, err
}

func (s *Store) UpsertAdmin(ctx context.Context, admin domain.Admin) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "username"}},
			DoUpdates: clause.AssignmentColumns([]string{"password", "updated_at"}),
		}).
		Create(&admin).Error
}

func (s *Store) EnsureAdminExists(ctx context.Context, username, password string) error {
	var count int64
	s.db.WithContext(ctx).Model(&domain.Admin{}).Count(&count)
	if count > 0 {
		return nil
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	admin := domain.Admin{
		Username: username,
		Password: string(hash),
	}
	return s.db.WithContext(ctx).Create(&admin).Error
}

// --- Categories ---

func (s *Store) AllCategories(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	err := s.db.WithContext(ctx).Order("name asc").Find(&categories).Error
	return categories, err
}

func (s *Store) CreateCategory(ctx context.Context, name string) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).
		Create(&domain.Category{Name: name}).Error
}

func (s *Store) DeleteCategory(ctx context.Context, id int) error {
	return s.db.WithContext(ctx).Delete(&domain.Category{}, id).Error
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
		KVKKAccepted:    input.KVKKAccepted,
		CustomerIP:      input.CustomerIP,
		Items:           input.Items,
		Status:          "Beklemede",
	}
	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

func (s *Store) UpdateOrder(ctx context.Context, id int, status, adminNote string) error {
	return s.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"admin_note": adminNote,
		}).Error
}

func (s *Store) AllOrders(ctx context.Context) ([]domain.Order, error) {
	var orders []domain.Order
	err := s.db.WithContext(ctx).
		Preload("Items").
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}
