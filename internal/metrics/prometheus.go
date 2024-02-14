package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type MetricsValues struct {
	sync.Mutex
	IntMetrics   map[string]int
	FloatMetrics map[string]float64
}

var (
	metricValues = MetricsValues{
		IntMetrics:   make(map[string]int),
		FloatMetrics: make(map[string]float64),
	}
)

// Update Metric FloatValue
func UpdateMetricValueInt(name string, value int) {
	metricValues.Lock()
	defer metricValues.Unlock()
	metricValues.IntMetrics[name] = value
}

// Update Metric FloatValue
func UpdateMetricValueFloat(name string, value float64) {
	metricValues.Lock()
	defer metricValues.Unlock()
	metricValues.FloatMetrics[name] = value
}

func Run() {
	UpdateMetricValueInt("warningsLastRun", 0)
	reg := prometheus.NewRegistry()
	if err := reg.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "schedule2slack",
			Subsystem:   "",
			Name:        "last_run_warnings",
			Help:        "Number of Warnings in last run.",
			ConstLabels: prometheus.Labels{"destination": "primary"},
		},
		func() float64 { return float64(metricValues.IntMetrics["warningsLastRun"]) },
	)); err != nil {
		log.Fatal(`GaugeFunc 'warningsLastRun' could not registered.`)
	}

	// make Prometheus client aware of our collector
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`ok`))
	})
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`ready`))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(
			`<html>
			<head><title>schedul2slack synchronziser</title></head>
			<body>
			<h1>schedule2slack</h1>
			<p><a href="/metrics">Metrics</a></p>
            <p><a href="/live">live probe</a></p>
            <p><a href="/ready">ready probe</a></p>
			<p><a href="https://github.com/sapcc/schedules2slack">Git Repository</a></p>
			</body>
			</html>`))
	})

	// https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	port := ":2112"
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
