package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"productsManager/internal/httpx"
	"productsManager/internal/products/api/handlers"
)

func NewProductsRouter(productHandler *handlers.Handler) http.Handler {
	router := chi.NewRouter()
	router.Post("/", productHandler.CreateProduct)
	router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		productHandler.DeleteProduct(w, r, chi.URLParam(r, "id"))
	})
	router.Get("/", productHandler.ListProducts)

	return router
}

func NewSystemRouter(registry *prometheus.Registry) http.Handler {
	router := chi.NewRouter()
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	return router
}
