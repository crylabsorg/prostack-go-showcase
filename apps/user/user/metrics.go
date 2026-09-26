package user

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	Created prometheus.Counter
	Fetched prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		Created: prometheus.NewCounter(prometheus.CounterOpts{Name: "user_created_total", Help: "Users created successfully."}),
		Fetched: prometheus.NewCounter(prometheus.CounterOpts{Name: "user_fetched_total", Help: "Users fetched successfully."}),
	}
	reg.MustRegister(m.Created, m.Fetched)
	return m
}
