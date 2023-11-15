package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer interface {
	ServeMetrics()
	IncrementEventCount(eventType string, responseCode int)
	ObserveRequestDuration(eventType string, responseCode int, latency float64)
}

func NewMetricsServer() MetricsServer {
	return &metricsServer{
		eventCounter:     make(map[string]prometheus.Counter),
		latencyHistogram: make(map[string]prometheus.Histogram),
	}
}

type metricsServer struct {
	eventCounter     map[string]prometheus.Counter
	latencyHistogram map[string]prometheus.Histogram
}

func (m *metricsServer) ServeMetrics() {
	fmt.Println("Starting up metric server")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

func (m *metricsServer) IncrementEventCount(eventType string, responseCode int) {
	var responseCodeClass string
	if responseCode >= 200 && responseCode < 300 {
		responseCodeClass = "2xx"
	} else {
		responseCodeClass = "5xx"
	}

	if m.eventCounter[eventType] == nil {
		m.eventCounter[eventType] = promauto.NewCounter(prometheus.CounterOpts{
			Name: "rabbitmq_ingress_event_count",
			Help: "Number of events received by the MQTT Broker",
			ConstLabels: prometheus.Labels{
				"broker_name":         "ktwin",
				"container_name":      "mqtt-dispatcher",
				"event_type":          eventType,
				"namespace_name":      "ktwin",
				"response_code":       fmt.Sprint(responseCode),
				"response_code_class": responseCodeClass,
				"unique_name":         "mqtt-dispatcher-pod",
			},
		})
	}

	m.eventCounter[eventType].Inc()
}

func (m *metricsServer) ObserveRequestDuration(eventType string, responseCode int, latency float64) {
	var responseCodeClass string
	if responseCode >= 200 && responseCode < 300 {
		responseCodeClass = "2xx"
	} else {
		responseCodeClass = "5xx"
	}

	if m.latencyHistogram[eventType] == nil {
		m.latencyHistogram[eventType] = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "rabbitmq_ingress_event_dispatch_latencies_bucket",
			Help:    "The time spent dispatching an event to a Channel.",
			Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
			ConstLabels: prometheus.Labels{
				"broker_name":         "ktwin",
				"container_name":      "mqtt-dispatcher",
				"event_type":          eventType,
				"namespace_name":      "ktwin",
				"response_code":       fmt.Sprint(responseCode),
				"response_code_class": responseCodeClass,
				"unique_name":         "mqtt-dispatcher-pod",
			},
		})
		prometheus.Register(m.latencyHistogram[eventType])
	}

	m.latencyHistogram[eventType].Observe(latency)

}
