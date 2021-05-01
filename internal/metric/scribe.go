package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	pusher *push.Pusher

	gauges            map[string]*prometheus.Gauge
	activeScribeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "caprice_scribe_active",
		Help: "a",
	})
)

func IncScribe(name string) {
	activeScribeGauge.Inc()
}

func DecScribe(name string) {
	activeScribeGauge.Dec()
}
