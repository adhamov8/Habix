package metrics

import (
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	ActiveChallengesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_challenges_total",
		Help: "Number of currently active challenges.",
	})

	CheckInsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "checkins_total",
		Help: "Total number of successful check-ins.",
	})

	DBConnectionsOpen = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_connections_open",
		Help: "Number of open database connections.",
	}, func() float64 { return 0 })
)

// RegisterDBStats replaces the default db_connections_open with one that reads from the given DB.
func RegisterDBStats(db *sqlx.DB) {
	prometheus.Unregister(DBConnectionsOpen)
	prometheus.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "db_connections_open",
		Help: "Number of open database connections.",
	}, func() float64 {
		return float64(db.Stats().OpenConnections)
	}))
}