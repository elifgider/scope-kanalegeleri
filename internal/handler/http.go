package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"kanalegeleri/go-app/internal/config"
	"kanalegeleri/go-app/internal/domain"
	"kanalegeleri/go-app/internal/repository"
	"kanalegeleri/go-app/internal/service"
)

type HTTPHandler struct {
	catalog       *service.CatalogService
	orders        *service.OrderService
	auth          *service.AuthService
	uploads       *service.UploadService
	homeTemplate  *template.Template
	adminTemplate *template.Template
	loginTemplate *template.Template
	uploadsDir    string
	telegram      config.TelegramConfig
	contactName   string
	contactPhone  string
	contactEmail  string
}

func NewHTTPHandler(
	cfg config.Settings,
	catalog *service.CatalogService,
	orders *service.OrderService,
	auth *service.AuthService,
	uploads *service.UploadService,
) (*HTTPHandler, error) {
	templatesDir := cfg.Paths.TemplatesDir

	homeTemplate, err := template.ParseFiles(filepath.Join(templatesDir, "index.html"))
	if err != nil {
		return nil, fmt.Errorf("parse home template: %w", err)
	}

	adminTemplate, err := template.ParseFiles(filepath.Join(templatesDir, "admin.html"))
	if err != nil {
		return nil, fmt.Errorf("parse admin template: %w", err)
	}

	loginTemplate, err := template.ParseFiles(filepath.Join(templatesDir, "admin_login.html"))
	if err != nil {
		return nil, fmt.Errorf("parse login template: %w", err)
	}

	return &HTTPHandler{
		catalog:       catalog,
		orders:        orders,
		auth:          auth,
		uploads:       uploads,
		homeTemplate:  homeTemplate,
		adminTemplate: adminTemplate,
		loginTemplate: loginTemplate,
		uploadsDir:    cfg.Paths.UploadsDir,
		telegram:      cfg.Telegram,
		contactName:   cfg.ContactName,
		contactPhone:  cfg.ContactPhone,
		contactEmail:  cfg.ContactEmail,
	}, nil
}

// requireAuth is a Gin middleware that checks admin session.
func (h *HTTPHandler) requireAuth(c *gin.Context) {
	if !h.auth.IsAuthenticated(c.Request) {
		c.Redirect(http.StatusSeeOther, "/admin/login")
		c.Abort()
		return
	}
	c.Next()
}

// --- API Handlers ---

func (h *HTTPHandler) handleProductsAPI(c *gin.Context) {
	products, err := h.catalog.ListProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ürünler alınamadı"})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *HTTPHandler) handleOrdersAPI(c *gin.Context) {
	var input struct {
		domain.CreateOrderRequest
		Website string `json:"website"` // Honeypot
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek gövdesi"})
		return
	}

	// Honeypot check: If website is not empty, it's a bot
	if input.Website != "" {
		log.Printf("Bot detected! Honeypot filled: %s", input.Website)
		c.JSON(http.StatusCreated, gin.H{"status": "ok"}) // Pretend it succeeded
		return
	}

	// Capture Client IP and set it
	input.CustomerIP = c.ClientIP()

	// Strict KVKK check
	if !input.KVKKAccepted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KVKK metnini onaylamanız gerekmektedir."})
		return
	}

	order, err := h.orders.CreateOrder(c.Request.Context(), input.CreateOrderRequest)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrder) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sipariş kaydedilemedi"})
		return
	}

	// Sipariş oluşturulurken OrderService içinde bildirim zaten gönderiliyor.
	c.JSON(http.StatusCreated, order)
}



// --- Page Handlers ---

func (h *HTTPHandler) handleHomePage(c *gin.Context) {
	products, err := h.catalog.ListProducts(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	productsJSON, err := json.Marshal(products)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	data := domain.HomePageData{
		Products:     template.JS(productsJSON),
		ContactName:  h.contactName,
		ContactPhone: h.contactPhone,
		ContactEmail: h.contactEmail,
	}

	c.Status(http.StatusOK)
	if err := h.homeTemplate.Execute(c.Writer, data); err != nil {
		log.Printf("home template error: %v", err)
	}
}

func (h *HTTPHandler) handleAdminPage(c *gin.Context) {
	products, err := h.catalog.ListProducts(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	orders, err := h.orders.ListOrders(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	categories, err := h.catalog.ListCategories(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if len(categories) == 0 {
		if err := h.syncCategoriesFromProducts(c.Request.Context(), products); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		categories, err = h.catalog.ListCategories(c.Request.Context())
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	message := c.Query("message")
	uploadedImageURL := strings.TrimSpace(c.Query("uploaded"))
	formProduct := domain.Product{ImageURL: uploadedImageURL}
	isEditing := false

	if editIDStr := strings.TrimSpace(c.Query("edit")); editIDStr != "" {
		editID, convErr := strconv.Atoi(editIDStr)
		if convErr == nil {
			product, found, getErr := h.catalog.GetProductByID(c.Request.Context(), editID)
			if getErr != nil {
				c.String(http.StatusInternalServerError, getErr.Error())
				return
			}
			if found {
				formProduct = product
				isEditing = true
				if uploadedImageURL != "" {
					formProduct.ImageURL = uploadedImageURL
				}
			}
		}
	}

	// Calculate Order Statistics
	stats := domain.OrderStats{Total: len(orders)}
	for _, o := range orders {
		switch o.Status {
		case "Beklemede":
			stats.Pending++
		case "Hazırlanıyor":
			stats.Processing++
		case "Kargolandı":
			stats.Shipped++
		case "Tamamlandı":
			stats.Completed++
		case "İptal":
			stats.Cancelled++
		}
	}

	c.Status(http.StatusOK)
	if err := h.adminTemplate.Execute(c.Writer, domain.AdminPageData{
		Products:         products,
		Orders:           orders,
		Categories:       categories,
		Message:          message,
		FormProduct:      formProduct,
		IsEditing:        isEditing,
		UploadedImageURL: uploadedImageURL,
		Stats:            stats,
	}); err != nil {
		log.Printf("admin template error: %v", err)
	}
}

func (h *HTTPHandler) syncCategoriesFromProducts(ctx context.Context, products []domain.Product) error {
	seen := make(map[string]struct{})

	for _, product := range products {
		name := strings.TrimSpace(product.Category)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		if err := h.catalog.CreateCategory(ctx, name); err != nil {
			return err
		}
	}

	return nil
}

func (h *HTTPHandler) handleAdminLogin(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		if h.auth.IsAuthenticated(c.Request) {
			c.Redirect(http.StatusSeeOther, "/admin")
			return
		}
		c.Status(http.StatusOK)
		if err := h.loginTemplate.Execute(c.Writer, nil); err != nil {
			log.Printf("login template error: %v", err)
		}
	case http.MethodPost:
		username := c.PostForm("username")
		password := c.PostForm("password")
		if !h.auth.Authenticate(username, password) {
			c.Redirect(http.StatusSeeOther, "/admin/login")
			return
		}
		h.auth.SetSession(c.Writer, username)
		c.Redirect(http.StatusSeeOther, "/admin")
	}
}

func (h *HTTPHandler) handleUpdateOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	status := c.PostForm("status")
	adminNote := c.PostForm("admin_note")

	if err := h.orders.UpdateOrder(c.Request.Context(), id, status, adminNote); err != nil {
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Siparis guncellenemedi", nil))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Siparis guncellendi", nil))
}

// --- Admin Product Handlers ---

func (h *HTTPHandler) handleCreateProduct(c *gin.Context) {
	price, _ := strconv.ParseFloat(strings.TrimSpace(c.PostForm("price")), 64)
	product := domain.Product{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Category:    c.PostForm("category"),
		ImageURL:    c.PostForm("image_url"),
		Price:       price,
	}

	if err := h.catalog.CreateProduct(c.Request.Context(), product); err != nil {
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Eksik urun bilgisi", nil))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Urun eklendi", nil))
}

func (h *HTTPHandler) handleUpdateProduct(c *gin.Context) {
	id, err := strconv.Atoi(strings.TrimSpace(c.PostForm("id")))
	if err != nil || id < 1 {
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Gecersiz urun id", nil))
		return
	}

	price, _ := strconv.ParseFloat(strings.TrimSpace(c.PostForm("price")), 64)
	product := domain.Product{
		ID:          uint(id),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Category:    c.PostForm("category"),
		ImageURL:    c.PostForm("image_url"),
		Price:       price,
	}

	if err := h.catalog.UpdateProduct(c.Request.Context(), product); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Urun bulunamadi", map[string]string{"edit": strconv.Itoa(id)}))
			return
		}
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Urun guncellenemedi", map[string]string{"edit": strconv.Itoa(id)}))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Urun guncellendi", nil))
}

func (h *HTTPHandler) handleDeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(strings.TrimSpace(c.PostForm("id")))
	if err != nil {
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Gecersiz urun id", nil))
		return
	}

	if err := h.catalog.DeleteProduct(c.Request.Context(), id); err != nil {
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Urun silinemedi", nil))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Urun silindi", nil))
}

func (h *HTTPHandler) handleCreateCategory(c *gin.Context) {
	name := c.PostForm("name")
	if err := h.catalog.CreateCategory(c.Request.Context(), name); err != nil {
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Kategori eklenemedi", nil))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Kategori eklendi", nil))
}

func (h *HTTPHandler) handleDeleteCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if err := h.catalog.DeleteCategory(c.Request.Context(), id); err != nil {
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Kategori silinemedi", nil))
		return
	}
	c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Kategori silindi", nil))
}

func (h *HTTPHandler) handleUploadImage(c *gin.Context) {
	imageURL, err := h.uploads.SaveImage(c.Request)
	if err != nil {
		log.Printf("upload error: %v", err)
		if expectsJSON(c.Request) {
			c.JSON(http.StatusBadRequest, gin.H{"error": uploadErrorMessage(err)})
			return
		}
		c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Gorsel yuklenemedi", nil))
		return
	}

	if expectsJSON(c.Request) {
		c.JSON(http.StatusOK, gin.H{
			"image_url": imageURL,
			"message":   "Gorsel yuklendi",
		})
		return
	}

	editID := strings.TrimSpace(c.PostForm("edit_id"))
	extra := map[string]string{"uploaded": imageURL}
	if editID != "" {
		extra["edit"] = editID
	}
	c.Redirect(http.StatusSeeOther, adminRedirectURL(c, "Gorsel yuklendi", extra))
}

func expectsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	requestedWith := r.Header.Get("X-Requested-With")
	return strings.Contains(accept, "application/json") || requestedWith == "XMLHttpRequest"
}

func uploadErrorMessage(err error) string {
	switch {
	case errors.Is(err, service.ErrUnsupportedImageType):
		return "Sadece JPG, PNG, WEBP veya GIF yukleyebilirsiniz"
	case errors.Is(err, service.ErrInvalidFileExtension):
		return "Dosya uzantisi anlasilamadi"
	default:
		return "Gorsel yuklenemedi"
	}
}

func adminRedirectURL(c *gin.Context, message string, extra map[string]string) string {
	values := url.Values{}
	values.Set("message", message)

	for key, value := range extra {
		if strings.TrimSpace(value) != "" {
			values.Set(key, value)
		}
	}

	redirectURL := "/admin"
	encoded := values.Encode()
	if encoded != "" {
		redirectURL += "?" + encoded
	}

	tab := strings.TrimSpace(c.PostForm("redirect_tab"))
	if tab != "" {
		redirectURL += "#" + tab
	}

	return redirectURL
}

func (h *HTTPHandler) handleLogout(c *gin.Context) {
	h.auth.ClearSession(c.Writer)
	c.Redirect(http.StatusSeeOther, "/admin/login")
}
