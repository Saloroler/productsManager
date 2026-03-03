package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"productsManager/internal/products/handlers"
	"productsManager/internal/products/metrics"
	"productsManager/internal/products/publisher"
	"productsManager/internal/products/store"
)

func Run(ctx context.Context, logger *slog.Logger, cfg Config) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	registry := prometheus.NewRegistry()
	productMetrics := metrics.New(registry)

	productStore := store.New(pool)
	var pub publisher.Publisher = publisher.NoopPublisher{}
	if cfg.SQSEnabled {
		client, queueURL, err := newSQSClient(ctx, cfg)
		if err != nil {
			return err
		}
		pub = publisher.NewSQS(client, queueURL)
	}

	router := handlers.New(productStore, pub, productMetrics).Routes()
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: time.Duration(cfg.ReadTimeout) * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("products service listening", "addr", server.Addr, "sqs_enabled", cfg.SQSEnabled)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
