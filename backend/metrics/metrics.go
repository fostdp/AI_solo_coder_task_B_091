package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "karez_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "karez_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	SensorDataReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "karez_sensor_data_received_total",
			Help: "Total number of sensor data received",
		},
	)

	SensorDataValid = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "karez_sensor_data_valid_total",
			Help: "Total number of valid sensor data",
		},
	)

	SensorDataInvalid = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "karez_sensor_data_invalid_total",
			Help: "Total number of invalid sensor data",
		},
	)

	SimulationsRun = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "karez_simulations_run_total",
			Help: "Total number of hydraulic simulations run",
		},
	)

	AllocationsRun = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "karez_allocations_run_total",
			Help: "Total number of water allocations run",
		},
	)

	AlertsTriggered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "karez_alerts_triggered_total",
			Help: "Total number of alerts triggered",
		},
		[]string{"alert_type", "level"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "karez_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"query_type"},
	)

	MQTTMessagesPublished = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "karez_mqtt_messages_published_total",
			Help: "Total number of MQTT messages published",
		},
	)

	MQTTMessagesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "karez_mqtt_messages_received_total",
			Help: "Total number of MQTT messages received",
		},
	)

	ChannelBackpressure = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "karez_channel_backpressure",
			Help: "Current channel backpressure level",
		},
		[]string{"channel_name"},
	)
)

func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

func ObserveHTTPRequest(method, path, status string) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
}

func ObserveHTTPDuration(method, path string, duration time.Duration) {
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func ObserveSensorData(valid bool) {
	SensorDataReceived.Inc()
	if valid {
		SensorDataValid.Inc()
	} else {
		SensorDataInvalid.Inc()
	}
}

func ObserveSimulationRun() {
	SimulationsRun.Inc()
}

func ObserveAllocationRun() {
	AllocationsRun.Inc()
}

func ObserveAlert(alertType, level string) {
	AlertsTriggered.WithLabelValues(alertType, level).Inc()
}

func ObserveDatabaseQuery(queryType string, duration time.Duration) {
	DatabaseQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

func ObserveMQTTPublish() {
	MQTTMessagesPublished.Inc()
}

func ObserveMQTTReceive() {
	MQTTMessagesReceived.Inc()
}

func SetChannelBackpressure(channelName string, value float64) {
	ChannelBackpressure.WithLabelValues(channelName).Set(value)
}
