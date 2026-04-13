package handler

import (
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
	var input domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek gövdesi"})
		return
	}

	order, err := h.orders.CreateOrder(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrder) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sipariş kaydedilemedi"})
		return
	}

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
		Products: template.JS(productsJSON),
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

	c.Status(http.StatusOK)
	if err := h.adminTemplate.Execute(c.Writer, domain.AdminPageData{
		Products:         products,
		Orders:           orders,
		Message:          message,
		FormProduct:      formProduct,
		IsEditing:        isEditing,
		UploadedImageURL: uploadedImageURL,
	}); err != nil {
		log.Printf("admin template error: %v", err)
	}
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
		h.auth.SetSession(c.Writer)
		c.Redirect(http.StatusSeeOther, "/admin")
	}
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
		c.Redirect(http.StatusSeeOther, "/admin?message=Eksik+urun+bilgisi")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin?message=Urun+eklendi")
}

func (h *HTTPHandler) handleUpdateProduct(c *gin.Context) {
	id, err := strconv.Atoi(strings.TrimSpace(c.PostForm("id")))
	if err != nil || id < 1 {
		c.Redirect(http.StatusSeeOther, "/admin?message=Gecersiz+urun+id")
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
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin?edit=%d&message=Urun+bulunamadi", id))
			return
		}
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin?edit=%d&message=Urun+guncellenemedi", id))
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin?message=Urun+guncellendi")
}

func (h *HTTPHandler) handleDeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(strings.TrimSpace(c.PostForm("id")))
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin?message=Gecersiz+urun+id")
		return
	}

	if err := h.catalog.DeleteProduct(c.Request.Context(), id); err != nil {
		c.Redirect(http.StatusSeeOther, "/admin?message=Urun+silinemedi")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin?message=Urun+silindi")
}

func (h *HTTPHandler) handleUploadImage(c *gin.Context) {
	imageURL, err := h.uploads.SaveImage(c.Request)
	if err != nil {
		log.Printf("upload error: %v", err)
		c.Redirect(http.StatusSeeOther, "/admin?message=Gorsel+yuklenemedi")
		return
	}

	editID := strings.TrimSpace(c.PostForm("edit_id"))
	redirectURL := fmt.Sprintf("/admin?message=Gorsel+yuklendi&uploaded=%s", url.QueryEscape(imageURL))
	if editID != "" {
		redirectURL += "&edit=" + url.QueryEscape(editID)
	}
	c.Redirect(http.StatusSeeOther, redirectURL)
}

func (h *HTTPHandler) handleLogout(c *gin.Context) {
	h.auth.ClearSession(c.Writer)
	c.Redirect(http.StatusSeeOther, "/admin/login")
}
