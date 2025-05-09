package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Счётчик загруженных файлов по типу
	filesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_bot_files_processed_total",
			Help: "Количество обработанных файлов по типу",
		},
		[]string{"type"},
	)

	// Время последней успешной обработки
	lastProcessedTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "telegram_bot_last_processed_timestamp",
		Help: "Время последней обработанной медиа (Unix timestamp)",
	})
)

func init() {
	// Регистрируем метрики
	prometheus.MustRegister(filesProcessed)
	prometheus.MustRegister(lastProcessedTimestamp)
}

// IncrementFileProcessed увеличивает счётчик для указанного типа
func IncrementFileProcessed(fileType string) {
	filesProcessed.WithLabelValues(fileType).Inc()
}

// UpdateLastProcessedTime обновляет временную метку
func UpdateLastProcessedTime() {
	lastProcessedTimestamp.SetToCurrentTime()
}

// HealthCheckHandler возвращает простой health check
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// StartMetricsServer запускает HTTP сервер с метриками
func StartMetricsServer(addr string) error {
	http.HandleFunc("/healthz", HealthCheckHandler)
	http.HandleFunc("/readyz", HealthCheckHandler)
	http.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(addr, nil)
}
