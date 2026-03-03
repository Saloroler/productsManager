package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"productsManager/internal/httpx"
	"productsManager/internal/products/metrics"
	"productsManager/internal/products/models"
	"productsManager/internal/products/publisher"
)

type Store interface {
	CreateProduct(ctx context.Context, name string, price int) (models.Product, error)
	DeleteProduct(ctx context.Context, id int64) (bool, error)
	ListProducts(ctx context.Context, page int, limit int) ([]models.Product, int64, error)
}

type Handler struct {
	store     Store
	publisher publisher.Publisher
	metrics   *metrics.Metrics
}

type createProductRequest struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type listProductsResponse struct {
	Items []models.Product `json:"items"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
	Total int64            `json:"total"`
}

func New(store Store, pub publisher.Publisher, productMetrics *metrics.Metrics) *Handler {
	return &Handler{
		store:     store,
		publisher: pub,
		metrics:   productMetrics,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/products", h.CreateProduct)
	r.Delete("/products/{id}", h.DeleteProduct)
	r.Get("/products", h.ListProducts)
	return r
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Price <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "price must be greater than zero")
		return
	}

	product, err := h.store.CreateProduct(r.Context(), req.Name, req.Price)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	if err := h.publisher.Publish(r.Context(), models.ProductEvent{
		Type:      "product.created",
		ProductID: product.ID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to publish product event")
		return
	}

	h.metrics.ProductsCreated.Inc()
	httpx.WriteJSON(w, http.StatusCreated, product)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	deleted, err := h.store.DeleteProduct(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete product")
		return
	}
	if !deleted {
		httpx.WriteError(w, http.StatusNotFound, "product not found")
		return
	}

	if err := h.publisher.Publish(r.Context(), models.ProductEvent{
		Type:      "product.deleted",
		ProductID: id,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to publish product event")
		return
	}

	h.metrics.ProductsDeleted.Inc()
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	page, limit, err := parsePagination(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, total, err := h.store.ListProducts(r.Context(), page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list products")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, listProductsResponse{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func parseProductID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid id")
	}

	return id, nil
}

func parsePagination(r *http.Request) (int, int, error) {
	page := 1
	limit := 20

	if raw := r.URL.Query().Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("page must be a positive integer")
		}
		page = parsed
	}

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("limit must be a positive integer")
		}
		limit = parsed
	}

	if limit > 100 {
		limit = 100
	}

	return page, limit, nil
}
