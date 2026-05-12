package middleware

import (
	"net/http"
	"strconv"
	"time"
	"tracker/internal/metrics"

	"github.com/go-chi/chi/v5"
)

// middleware, который пишет в Prometheus метрики по каждому запросу
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		// Берём шаблон маршрута из chi, чтобы пути с параметрами сворачивались в одну метрику
		path := r.URL.Path
		if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
			path = rctx.RoutePattern()
		}

		status := strconv.Itoa(sw.status)
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}
