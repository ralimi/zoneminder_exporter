package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct{}

func NewExporter() *Exporter {
	return &Exporter{}
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
}
