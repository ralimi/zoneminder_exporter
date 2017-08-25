package exporter

import (
	"context"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/ralimi/zoneminder_exporter/zoneminder"
)

const (
	eventLookbackDuration = 3 * time.Hour
)

type Exporter struct {
	zmClient               zoneminder.Client
	collectTimeout         time.Duration
	lastEventStartTimeDesc *prometheus.Desc
	lastEventEndTimeDesc   *prometheus.Desc
	daemonRunningDesc      *prometheus.Desc
	monitorConfiguredDesc  *prometheus.Desc
}

func New(zmApiUrl string, collectTimeout time.Duration) *Exporter {
	return &Exporter{
		zmClient:       zoneminder.New(zmApiUrl),
		collectTimeout: collectTimeout,
		lastEventStartTimeDesc: prometheus.NewDesc(
			"zoneminder_last_event_start_time",
			"Start time of last event",
			[]string{"monitor"},
			prometheus.Labels{},
		),
		lastEventEndTimeDesc: prometheus.NewDesc(
			"zoneminder_last_event_end_time",
			"End time of last event",
			[]string{"monitor"},
			prometheus.Labels{},
		),
		daemonRunningDesc: prometheus.NewDesc(
			"zoneminder_daemon_running",
			"Status of the ZoneMinder daemon",
			[]string{},
			prometheus.Labels{},
		),
		monitorConfiguredDesc: prometheus.NewDesc(
			"zoneminder_monitor_configured",
			"Monitor configured in ZoneMinder",
			[]string{"monitor"},
			prometheus.Labels{},
		),
	}
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.lastEventStartTimeDesc
	ch <- e.lastEventEndTimeDesc
	ch <- e.daemonRunningDesc
	ch <- e.monitorConfiguredDesc
}

func groupByMonitor(events []zoneminder.Event) map[string][]zoneminder.Event {
	result := make(map[string][]zoneminder.Event)
	for _, e := range events {
		result[e.Monitor.Name] = append(result[e.Monitor.Name], e)
	}
	return result
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), e.collectTimeout)
	defer cancel()

	// Daemon status metrics
	if running, err := e.zmClient.DaemonRunning(ctx); err == nil {
		runningInt := 1
		if !running {
			runningInt = 0
		}
		ch <- prometheus.MustNewConstMetric(
			e.daemonRunningDesc,
			prometheus.GaugeValue,
			float64(runningInt),
		)
	} else {
		log.Errorf("Failed to check if ZoneMinder was running: %v", err)
	}

	// Event metrics
	minStart := time.Now().Add(-1 * eventLookbackDuration)
	if events, err := e.zmClient.Events(ctx, minStart); err == nil {
		// Export metrics for each monitor
		for monitor, mEvents := range groupByMonitor(events) {
			// Find the last event
			sort.Slice(mEvents, func(i, j int) bool { return mEvents[i].Start.Before(mEvents[j].Start) })
			last := mEvents[len(mEvents)-1]
			// Export metrics for the event
			ch <- prometheus.MustNewConstMetric(
				e.lastEventStartTimeDesc,
				prometheus.GaugeValue,
				float64(last.Start.Unix()),
				monitor,
			)
			ch <- prometheus.MustNewConstMetric(
				e.lastEventEndTimeDesc,
				prometheus.GaugeValue,
				float64(last.End.Unix()),
				monitor,
			)
		}
	} else {
		log.Errorf("Failed to fetch ZoneMinder events: %v", err)
	}

	// Configured monitors
	if monitors, err := e.zmClient.Monitors(ctx); err == nil {
		for _, m := range monitors {
			ch <- prometheus.MustNewConstMetric(
				e.monitorConfiguredDesc,
				prometheus.GaugeValue,
				float64(1),
				m.Name,
			)
		}
	} else {
		log.Errorf("Failed to fetch ZoneMinder monitors: %v", err)
	}
}
