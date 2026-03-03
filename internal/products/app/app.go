package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"

	sharedaws "productsManager/internal/aws"
	"productsManager/internal/products/api/handlers"
	"productsManager/internal/products/api/routes"
	"productsManager/internal/products/metrics"
	"productsManager/internal/products/publisher"
	"productsManager/internal/products/repo"
)

func Run(ctx context.Context, logger *slog.Logger, cfg Config) error {
	conn, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	registry := prometheus.NewRegistry()
	productRepo := repo.NewProductRepo(conn)
	productPublisher, err := newPublisher(ctx, cfg)
	if err != nil {
		return err
	}
	productMetrics := metrics.New(registry)
	productHandler := handlers.New(productRepo, productPublisher, productMetrics)

	router := chi.NewRouter()
	router.Mount("/products", routes.NewProductsRouter(productHandler))
	router.Mount("/", routes.NewSystemRouter(registry))

	listener := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: time.Duration(cfg.ReadTimeout) * time.Second,
	}

	stopShutdown := context.AfterFunc(ctx, func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
		defer cancel()
		_ = listener.Shutdown(shutdownCtx)
	})
	defer stopShutdown()

	logger.Info("products service listening", "addr", listener.Addr, "sqs_enabled", cfg.SQSEnabled)

	if err := listener.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func newPublisher(ctx context.Context, cfg Config) (publisher.Publisher, error) {
	if !cfg.SQSEnabled {
		return publisher.NoopPublisher{}, nil
	}

	client, queueURL, err := sharedaws.NewSQSClient(ctx, sharedaws.SQSConfig{
		Region:          cfg.AWSRegion,
		EndpointURL:     cfg.AWSEndpointURL,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretKey,
		QueueName:       cfg.SQSQueueName,
		QueueURL:        cfg.SQSQueueURL,
	})
	if err != nil {
		return nil, err
	}

	return publisher.NewSQS(client, queueURL), nil
}
