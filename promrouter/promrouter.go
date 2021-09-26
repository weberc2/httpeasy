package promrouter

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	pz "github.com/weberc2/httpeasy"
)

// PromRouter allows registering routes with Prometheus monitoring out of the
// box.
type PromRouter struct {
	// Durations is a HistogramVec. PromRouter will make observations attaching
	// labels for the request path and the response HTTP status, so Durations
	// should have two labels (preferably `"path"` and `"status"`).
	Durations *prometheus.HistogramVec

	// Router is the underlying `httpeasy.Router` instance.
	*pz.Router
}

// NewWithDefaults creates a PromRouter with default values.
func NewWithDefaults() *PromRouter {
	return &PromRouter{
		Durations: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Buckets: prometheus.ExponentialBuckets(0.000001, 10, 8),
			},
			[]string{"path", "status"},
		),
		Router: pz.NewRouter(),
	}
}

// Register registers routes with the PromRouter and returns the PromRouter
// instance.
func (pm *PromRouter) Register(log pz.LogFunc, routes ...pz.Route) *PromRouter {
	for i, route := range routes {
		routes[i].Handler = func(r pz.Request) pz.Response {
			start := time.Now()
			rsp := route.Handler(r)
			pm.Durations.WithLabelValues(
				route.Path,
				strconv.Itoa(rsp.Status),
			).Observe(time.Since(start).Seconds())
			return rsp
		}
	}
	pm.Router.Register(log, routes...)
	return pm
}
