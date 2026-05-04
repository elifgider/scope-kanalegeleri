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
	UpdateOrder(ctx context.Context, id int, status, adminNote string) error
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
		if notifyErr := s.notifier.Notify(order); notifyErr != nil {
			fmt.Printf("⚠️ Telegram bildirim hatasi: %v\n", notifyErr)
		}
	}

	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context) ([]domain.Order, error) {
	return s.repo.AllOrders(ctx)
}

func (s *OrderService) UpdateOrder(ctx context.Context, id int, status, adminNote string) error {
	if id < 1 {
		return ErrInvalidOrderID
	}
	status = strings.TrimSpace(status)
	adminNote = strings.TrimSpace(adminNote)
	return s.repo.UpdateOrder(ctx, id, status, adminNote)
}

type TelegramNotifier struct {
	BotToken string
	ChatID   string
}

func (n TelegramNotifier) Notify(order domain.Order) error {
	if strings.TrimSpace(n.BotToken) == "" || strings.TrimSpace(n.ChatID) == "" {
		fmt.Println("ℹ️ Telegram bildirim ayarları boş (BotToken veya ChatID eksik), bildirim gönderilmedi.")
		return nil
	}

	lines := []string{
		"<b>🔔 YENİ SİPARİŞ / TEKLİF</b>",
		"---------------------------",
		fmt.Sprintf("<b>Tip:</b> %s", order.FormType),
		fmt.Sprintf("<b>Müşteri:</b> %s", order.CustomerName),
		fmt.Sprintf("<b>Telefon:</b> %s", order.CustomerPhone),
		fmt.Sprintf("<b>Şehir/İlçe:</b> %s", order.CustomerAddress),
	}

	if order.CustomerEmail != "" && !strings.Contains(order.CustomerEmail, "@example.com") {
		lines = append(lines, fmt.Sprintf("<b>E-posta:</b> %s", order.CustomerEmail))
	}

	if order.FullAddress != "" {
		lines = append(lines, fmt.Sprintf("<b>Adres:</b> <i>%s</i>", order.FullAddress))
	}
	if order.Note != "" {
		lines = append(lines, fmt.Sprintf("<b>Not:</b> %s", order.Note))
	}

	if len(order.Items) > 0 {
		lines = append(lines, "", "<b>📦 Ürünler:</b>")
		for _, item := range order.Items {
			lines = append(lines, fmt.Sprintf("- %s (x%d)", item.ProductName, item.Quantity))
		}
	}

	response, err := http.PostForm(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.BotToken),
		url.Values{
			"chat_id":    {n.ChatID},
			"text":       {strings.Join(lines, "\n")},
			"parse_mode": {"HTML"},
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
