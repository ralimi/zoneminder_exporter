package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	"github.com/ralimi/zoneminder_exporter/exporter"
)

var (
	showVersion = flag.Bool(
		"version", false,
		"Print version information.",
	)
	listenAddress = flag.String(
		"web.listen-address", ":9180",
		"Address to listen on for web interface and telemetry.",
	)
	metricPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose metrics.",
	)
)

const (
	namespace = "zoneminder"
)

func init() {
	prometheus.MustRegister(version.NewCollector("zoneminder_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("zoneminder_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting zoneminder_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exp := exporter.NewExporter()
	prometheus.MustRegister(exp)

	http.Handle(*metricPath, promhttp.Handler())

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
