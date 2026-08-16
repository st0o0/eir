package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	eventsReceived *prometheus.CounterVec
	healsTotal     *prometheus.CounterVec
	healDuration   *prometheus.HistogramVec
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		eventsReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "eir_events_received_total",
			Help: "Total number of Docker events received.",
		}, []string{"action"}),
		healsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "eir_heals_total",
			Help: "Total number of healing operations.",
		}, []string{"master", "status"}),
		healDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "eir_heal_duration_seconds",
			Help:    "Duration of healing operations in seconds.",
			Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60},
		}, []string{"master"}),
	}
	reg.MustRegister(m.eventsReceived, m.healsTotal, m.healDuration)
	return m
}

func (m *Metrics) RecordEvent(action string) {
	m.eventsReceived.WithLabelValues(action).Inc()
}

func (m *Metrics) RecordHeal(master string, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "fail"
	}
	m.healsTotal.WithLabelValues(master, status).Inc()
	m.healDuration.WithLabelValues(master).Observe(duration.Seconds())
}
