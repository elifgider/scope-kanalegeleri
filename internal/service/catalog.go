package service

import (
	"context"
	"strings"

	"kanalegeleri/go-app/internal/domain"
)

type CatalogRepository interface {
	AllProducts(ctx context.Context) ([]domain.Product, error)
	GetProductByID(ctx context.Context, id int) (domain.Product, bool, error)
	CreateProduct(ctx context.Context, product domain.Product) error
	UpdateProduct(ctx context.Context, product domain.Product) error
	DeleteProduct(ctx context.Context, id int) error
}

type CatalogService struct {
	repo CatalogRepository
}

func NewCatalogService(repo CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return s.repo.AllProducts(ctx)
}

func (s *CatalogService) GetProductByID(ctx context.Context, id int) (domain.Product, bool, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *CatalogService) CreateProduct(ctx context.Context, product domain.Product) error {
	product.Name = strings.TrimSpace(product.Name)
	product.Description = strings.TrimSpace(product.Description)
	product.Category = strings.TrimSpace(product.Category)
	product.ImageURL = strings.TrimSpace(product.ImageURL)

	if product.Name == "" || product.Category == "" || product.ImageURL == "" {
		return ErrInvalidProduct
	}

	return s.repo.CreateProduct(ctx, product)
}

func (s *CatalogService) UpdateProduct(ctx context.Context, product domain.Product) error {
	product.Name = strings.TrimSpace(product.Name)
	product.Description = strings.TrimSpace(product.Description)
	product.Category = strings.TrimSpace(product.Category)
	product.ImageURL = strings.TrimSpace(product.ImageURL)

	if product.ID < 1 || product.Name == "" || product.Category == "" || product.ImageURL == "" {
		return ErrInvalidProduct
	}

	return s.repo.UpdateProduct(ctx, product)
}

func (s *CatalogService) DeleteProduct(ctx context.Context, id int) error {
	if id < 1 {
		return ErrInvalidProductID
	}
	return s.repo.DeleteProduct(ctx, id)
}
