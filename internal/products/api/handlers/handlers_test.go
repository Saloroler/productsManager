package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"

	"productsManager/internal/products/api/handlers"
	"productsManager/internal/products/api/routes"
	"productsManager/internal/products/metrics"
	"productsManager/internal/products/models"
	"productsManager/internal/products/publisher"
	"productsManager/internal/products/repo"
	"productsManager/internal/testutil"
)

func TestCreateProduct(t *testing.T) {
	productRepo, router := newTestRouter(t)

	body := bytes.NewBufferString(`{"name":"Keyboard","price":12345}`)
	req := httptest.NewRequest(http.MethodPost, "/products", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var got models.Product
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != 1 {
		t.Fatalf("expected id 1, got %d", got.ID)
	}
	if got.Name != "Keyboard" {
		t.Fatalf("expected name Keyboard, got %s", got.Name)
	}
	if got.Price != 12345 {
		t.Fatalf("expected price 12345, got %d", got.Price)
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set")
	}

	items, total, err := productRepo.ListProducts(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("list products from repo: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one product in database, total=%d items=%d", total, len(items))
	}
}

func TestListProductsPagination(t *testing.T) {
	productRepo, router := newTestRouter(t)
	ctx := context.Background()

	first, err := productRepo.CreateProduct(ctx, "First", 100)
	if err != nil {
		t.Fatalf("seed first product: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	second, err := productRepo.CreateProduct(ctx, "Second", 200)
	if err != nil {
		t.Fatalf("seed second product: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	third, err := productRepo.CreateProduct(ctx, "Third", 300)
	if err != nil {
		t.Fatalf("seed third product: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/products?page=1&limit=2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var pageOne struct {
		Items []models.Product `json:"items"`
		Page  int              `json:"page"`
		Limit int              `json:"limit"`
		Total int64            `json:"total"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&pageOne); err != nil {
		t.Fatalf("decode page one: %v", err)
	}

	if pageOne.Page != 1 || pageOne.Limit != 2 || pageOne.Total != 3 {
		t.Fatalf("unexpected page one metadata: %+v", pageOne)
	}
	if len(pageOne.Items) != 2 {
		t.Fatalf("expected 2 items on first page, got %d", len(pageOne.Items))
	}
	if pageOne.Items[0].ID != third.ID || pageOne.Items[1].ID != second.ID {
		t.Fatalf("unexpected order on first page: %+v", pageOne.Items)
	}

	req = httptest.NewRequest(http.MethodGet, "/products?page=2&limit=2", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var pageTwo struct {
		Items []models.Product `json:"items"`
		Page  int              `json:"page"`
		Limit int              `json:"limit"`
		Total int64            `json:"total"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&pageTwo); err != nil {
		t.Fatalf("decode page two: %v", err)
	}

	if pageTwo.Page != 2 || pageTwo.Limit != 2 || pageTwo.Total != 3 {
		t.Fatalf("unexpected page two metadata: %+v", pageTwo)
	}
	if len(pageTwo.Items) != 1 || pageTwo.Items[0].ID != first.ID {
		t.Fatalf("unexpected page two items: %+v", pageTwo.Items)
	}
}

func TestDeleteProduct(t *testing.T) {
	productRepo, router := newTestRouter(t)
	product, err := productRepo.CreateProduct(context.Background(), "Mouse", 5000)
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/products/"+strconv.FormatInt(product.ID, 10), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	items, total, err := productRepo.ListProducts(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("list products after delete: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("expected empty database after delete, total=%d items=%d", total, len(items))
	}

	req = httptest.NewRequest(http.MethodDelete, "/products/"+strconv.FormatInt(product.ID, 10), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func newTestRouter(t *testing.T) (*repo.ProductRepo, http.Handler) {
	t.Helper()

	conn := testutil.OpenTestConn(t)
	productRepo := repo.NewProductRepo(conn)
	if err := productRepo.TruncateProducts(context.Background()); err != nil {
		t.Fatalf("truncate products: %v", err)
	}

	registry := prometheus.NewRegistry()
	productMetrics := metrics.New(registry)
	productHandler := handlers.New(productRepo, publisher.NoopPublisher{}, productMetrics)

	router := chi.NewRouter()
	router.Mount("/products", routes.NewProductsRouter(productHandler))
	router.Mount("/", routes.NewSystemRouter(registry))

	return productRepo, router
}
