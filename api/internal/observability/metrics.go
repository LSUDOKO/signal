package observability

import (
	"log/slog"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for Signal.
type Metrics struct {
	EventsReceived      *prometheus.CounterVec
	MessagesProcessed   *prometheus.CounterVec
	AICallsMade         *prometheus.CounterVec
	AICallDuration      *prometheus.HistogramVec
	FocusModeTriggers   prometheus.Counter
	TranslationsSent    prometheus.Counter
	CatchUpSearches     prometheus.Counter
	DigestsSent         prometheus.Counter
	DeepWorkSessions    prometheus.Counter
	Errors              *prometheus.CounterVec
	ActiveUsers         prometheus.Gauge
	HTTPRequests        *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		EventsReceived: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "signal_events_received_total",
				Help: "Total number of Slack events received by type",
			},
			[]string{"type"},
		),
		MessagesProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "signal_messages_processed_total",
				Help: "Total number of messages processed",
			},
			[]string{"channel_type"},
		),
		AICallsMade: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "signal_ai_calls_total",
				Help: "Total number of AI API calls",
			},
			[]string{"feature"},
		),
		AICallDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "signal_ai_call_duration_seconds",
				Help:    "Duration of AI API calls",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"feature"},
		),
		FocusModeTriggers: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "signal_focus_mode_triggers_total",
				Help: "Total number of focus mode triggers",
			},
		),
		TranslationsSent: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "signal_translations_sent_total",
				Help: "Total number of social translations sent",
			},
		),
		CatchUpSearches: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "signal_catchup_searches_total",
				Help: "Total number of catch-up searches performed",
			},
		),
		DigestsSent: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "signal_digests_sent_total",
				Help: "Total number of digests sent",
			},
		),
		DeepWorkSessions: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "signal_deep_work_sessions_total",
				Help: "Total number of deep work sessions started",
			},
		),
		Errors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "signal_errors_total",
				Help: "Total number of errors by type",
			},
			[]string{"feature", "error_type"},
		),
		ActiveUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "signal_active_users",
				Help: "Number of active users",
			},
		),
		HTTPRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "signal_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "signal_http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
}

// SetupLogging configures structured JSON logging via slog.
func SetupLogging(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", time.Now().Format(time.RFC3339))
			}
			return a
		},
	})

	slog.SetDefault(slog.New(handler))
	slog.Info("logging initialized", "level", level)
}

// TrackAICall is a helper to measure AI call duration.
func TrackAICall(metrics *Metrics, feature string, start time.Time, err error) {
	duration := time.Since(start)
	metrics.AICallsMade.WithLabelValues(feature).Inc()
	metrics.AICallDuration.WithLabelValues(feature).Observe(duration.Seconds())

	if err != nil {
		metrics.Errors.WithLabelValues(feature, "ai_error").Inc()
		slog.Error("ai call failed", "feature", feature, "error", err, "duration", duration)
	} else {
		slog.Debug("ai call completed", "feature", feature, "duration", duration)
	}
}
