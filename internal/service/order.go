package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"kanalegeleri/go-app/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, input domain.CreateOrderRequest) (domain.Order, error)
	AllOrders(ctx context.Context) ([]domain.Order, error)
}

type OrderNotifier interface {
	Notify(order domain.Order) error
}

type OrderService struct {
	repo     OrderRepository
	notifier OrderNotifier
}

func NewOrderService(repo OrderRepository, notifier OrderNotifier) *OrderService {
	return &OrderService{repo: repo, notifier: notifier}
}

func (s *OrderService) CreateOrder(ctx context.Context, input domain.CreateOrderRequest) (domain.Order, error) {
	input.FormType = strings.TrimSpace(input.FormType)
	input.CustomerName = strings.TrimSpace(input.CustomerName)
	input.CustomerEmail = strings.TrimSpace(input.CustomerEmail)
	input.CustomerPhone = strings.TrimSpace(input.CustomerPhone)
	input.CustomerAddress = strings.TrimSpace(input.CustomerAddress)
	input.FullAddress = strings.TrimSpace(input.FullAddress)
	input.Note = strings.TrimSpace(input.Note)

	if err := ValidateOrder(input); err != nil {
		return domain.Order{}, errors.Join(ErrInvalidOrder, err)
	}

	order, err := s.repo.CreateOrder(ctx, input)
	if err != nil {
		return domain.Order{}, err
	}

	if s.notifier != nil {
		_ = s.notifier.Notify(order)
	}

	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context) ([]domain.Order, error) {
	return s.repo.AllOrders(ctx)
}

type TelegramNotifier struct {
	BotToken string
	ChatID   string
}

func (n TelegramNotifier) Notify(order domain.Order) error {
	if strings.TrimSpace(n.BotToken) == "" || strings.TrimSpace(n.ChatID) == "" {
		return nil
	}

	lines := []string{
		"Yeni siparis/teklif geldi",
		"Tip: " + order.FormType,
		"Musteri: " + order.CustomerName,
		"Telefon: " + order.CustomerPhone,
		"E-posta: " + order.CustomerEmail,
		"Sehir/Ilce: " + order.CustomerAddress,
	}
	if order.FullAddress != "" {
		lines = append(lines, "Adres: "+order.FullAddress)
	}
	if order.Note != "" {
		lines = append(lines, "Not: "+order.Note)
	}
	lines = append(lines, "Urunler:")
	for _, item := range order.Items {
		lines = append(lines, fmt.Sprintf("- %s x%d", item.ProductName, item.Quantity))
	}

	response, err := http.PostForm(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.BotToken),
		url.Values{
			"chat_id": {n.ChatID},
			"text":    {strings.Join(lines, "\n")},
		},
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("telegram status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
