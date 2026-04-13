package domain

import (
	"html/template"
	"time"
)

type Product struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"                     json:"id"`
	Name        string    `gorm:"not null"                                     json:"name"`
	Description string    `gorm:"not null;default:''"                          json:"description"`
	Category    string    `gorm:"not null"                                     json:"category"`
	ImageURL    string    `gorm:"not null;column:image_url"                    json:"image_url"`
	Price       float64   `gorm:"not null;default:0;type:numeric(12,2)"        json:"price"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type OrderItem struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"                    json:"-"`
	OrderID     uint   `gorm:"not null;index"                              json:"-"`
	ProductID   int    `gorm:"not null"                                    json:"productId"`
	ProductName string `gorm:"not null"                                    json:"productName"`
	Quantity    int    `gorm:"not null;check:quantity > 0"                 json:"quantity"`
}

type Order struct {
	ID              uint        `gorm:"primaryKey;autoIncrement"                     json:"id"`
	FormType        string      `gorm:"not null"                                     json:"formType"`
	CustomerName    string      `gorm:"not null"                                     json:"customerName"`
	CustomerEmail   string      `gorm:"not null"                                     json:"customerEmail"`
	CustomerPhone   string      `gorm:"not null"                                     json:"customerPhone"`
	CustomerAddress string      `gorm:"not null"                                     json:"customerAddress"`
	FullAddress     string      `gorm:"not null;default:''"                          json:"fullAddress"`
	Note            string      `gorm:"not null;default:''"                          json:"note"`
	Items           []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt       time.Time   `json:"createdAt"`
}

// CreateOrderRequest HTTP katmanından gelen ham istek — DB modeli değil.
type CreateOrderRequest struct {
	FormType        string      `json:"formType"`
	CustomerName    string      `json:"customerName"`
	CustomerEmail   string      `json:"customerEmail"`
	CustomerPhone   string      `json:"customerPhone"`
	CustomerAddress string      `json:"customerAddress"`
	FullAddress     string      `json:"fullAddress"`
	Note            string      `json:"note"`
	Items           []OrderItem `json:"items"`
}

type HomePageData struct {
	Products template.JS
}

type AdminPageData struct {
	Products         []Product
	Orders           []Order
	Message          string
	FormProduct      Product
	IsEditing        bool
	UploadedImageURL string
}
