package service

import (
	"errors"
	"strings"

	"kanalegeleri/go-app/internal/domain"
)

func ValidateOrder(input domain.CreateOrderRequest) error {
	if strings.TrimSpace(input.CustomerName) == "" {
		return errors.New("Ad Soyad zorunludur")
	}
	if strings.TrimSpace(input.CustomerEmail) == "" {
		return errors.New("E-posta zorunludur")
	}
	if strings.TrimSpace(input.CustomerPhone) == "" {
		return errors.New("Telefon zorunludur")
	}
	if strings.TrimSpace(input.CustomerAddress) == "" {
		return errors.New("Şehir / İlçe zorunludur")
	}
	if len(input.Items) == 0 {
		return errors.New("En az bir ürün seçmelisiniz")
	}
	for _, item := range input.Items {
		if item.ProductID == 0 || strings.TrimSpace(item.ProductName) == "" || item.Quantity < 1 {
			return errors.New("Sipariş kalemleri geçersiz")
		}
	}
	return nil
}
