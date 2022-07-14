package writeproxy

import (
	"time"

	"github.com/weaveworks/common/instrument"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefix = "graphite_proxy_ingester"
)

//go:generate mockery --inpackage --testonly --case underscore --name Recorder
type Recorder interface {
	measureReceivedSamples(user string, count int)
	measureIncomingSamples(user string, count int)
	measureRejectedSamples(user, reason string)
	measureConversionDuration(user string, duration time.Duration)
}

// NewRecorder returns a new Prometheus metrics Recorder.
// It ensures that the graphite ingester metrics are properly registered.
func NewRecorder(reg prometheus.Registerer) Recorder {
	r := &prometheusRecorder{
		receivedSamples: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "received_samples_total",
			Help:      "The total number of received samples, excluding rejected and deduped samples.",
		}, []string{"user"}),
		incomingSamples: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "samples_in_total",
			Help: "The total number of samples that have come in to the graphite write proxy, including rejected " +
				"or deduped samples.",
		}, []string{"user"}),
		rejectedSamples: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "rejected_samples_total",
			Help:      "The total number of samples that were rejected.",
		}, []string{"user", "reason"}),
		conversionDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: prefix,
			Name:      "data_conversion_seconds",
			Help:      "Time (in seconds) spent converting ingested Graphite data into Prometheus data.",
			Buckets:   instrument.DefBuckets,
		}, []string{"user"}),
	}

	reg.MustRegister(r.receivedSamples, r.incomingSamples, r.rejectedSamples, r.conversionDuration)

	return r
}

// prometheusRecorder knows the metrics of the ingester and how to measure them for
// Prometheus.
type prometheusRecorder struct {
	receivedSamples    *prometheus.CounterVec
	incomingSamples    *prometheus.CounterVec
	rejectedSamples    *prometheus.CounterVec
	conversionDuration *prometheus.HistogramVec
}

// measureMetricsParsed measures the total amount of received samples on Prometheus.
func (r prometheusRecorder) measureReceivedSamples(user string, count int) {
	r.receivedSamples.WithLabelValues(user).Add(float64(count))
}

// measureIncomingSamples measures the total amount of incoming samples on Prometheus.
func (r prometheusRecorder) measureIncomingSamples(user string, count int) {
	r.incomingSamples.WithLabelValues(user).Add(float64(count))
}

// measureRejectedSamples measures the total amount of rejected samples on Prometheus.
func (r prometheusRecorder) measureRejectedSamples(user, reason string) {
	r.rejectedSamples.WithLabelValues(user, reason).Add(1)
}

// measureConversionDuration measures the total time spent translating samples to Prometheus format
func (r prometheusRecorder) measureConversionDuration(user string, duration time.Duration) {
	r.conversionDuration.WithLabelValues(user).Observe(duration.Seconds())
}