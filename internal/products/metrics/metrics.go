package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	ProductsCreated prometheus.Counter
	ProductsDeleted prometheus.Counter
}

func New(registry prometheus.Registerer) *Metrics {
	created := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "products_created_total",
		Help: "Total number of created products.",
	})
	deleted := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "products_deleted_total",
		Help: "Total number of deleted products.",
	})

	registry.MustRegister(created, deleted)

	return &Metrics{
		ProductsCreated: created,
		ProductsDeleted: deleted,
	}
}
