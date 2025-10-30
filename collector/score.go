package collector

import "github.com/prometheus/client_golang/prometheus"

type scoreCollector struct {
	utilizationScore *prometheus.Desc
	healthScore      *prometheus.Desc
}

func NewScoreCollector() *scoreCollector {
	return &scoreCollector{
		utilizationScore: prometheus.NewDesc(
			"syscore_utilization_score",
			"Measures weigthed system utilization",
			nil,
			nil,
		),
		healthScore: prometheus.NewDesc(
			"syscore_health_score",
			"Health score",
			nil,
			nil,
		),
	}
}

func (sc *scoreCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(sc, ch)
}

func (sc *scoreCollector) Collect(ch chan<- prometheus.Metric) {
	println("Hello")
}
