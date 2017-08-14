package zoneminder

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/common/log"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

type Monitor struct {
	Id   string
	Name string
}

type Event struct {
	Id      string
	Name    string
	Start   time.Time
	End     time.Time
	Monitor *Monitor
}

type Client interface {
	DaemonRunning(ctx context.Context) (bool, error)
	Monitors(ctx context.Context) ([]Monitor, error)
	Events(ctx context.Context, minStart time.Time) ([]Event, error)
}

type client struct {
	apiUrl string
	cli    *http.Client
}

func New(apiUrl string) Client {
	return &client{
		apiUrl: apiUrl,
		cli:    &http.Client{},
	}
}

func (c *client) doGet(ctx context.Context, subPath string, result interface{}) error {
	url := fmt.Sprintf("%s/%s", c.apiUrl, subPath)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Failed to create request: %v", err)
	}

	req = req.WithContext(ctx)
	resp, err := c.cli.Do(req)
	if err != nil {
		return fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read request: %v", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("Failed to parse response: %v", err)
	}

	return nil
}

func (c *client) DaemonRunning(ctx context.Context) (bool, error) {
	type rsp struct {
		Result int64
	}

	var r rsp
	if err := c.doGet(ctx, "host/daemonCheck.json", &r); err != nil {
		return false, fmt.Errorf("Request failed: %v", err)
	}

	return r.Result != 0, nil
}

func (c *client) Monitors(ctx context.Context) ([]Monitor, error) {
	type rspMonitor struct {
		Id   string
		Name string
	}

	type rspMonitorItem struct {
		Monitor *rspMonitor
	}

	type rsp struct {
		Monitors []*rspMonitorItem
	}

	var r rsp
	if err := c.doGet(ctx, "monitors.json", &r); err != nil {
		return nil, fmt.Errorf("Request failed: %v", err)
	}

	var result []Monitor
	for _, mi := range r.Monitors {
		if mi.Monitor != nil {
			result = append(result, Monitor{
				Id:   mi.Monitor.Id,
				Name: mi.Monitor.Name,
			})
		}
	}
	return result, nil
}

func (c *client) Events(ctx context.Context, minStart time.Time) ([]Event, error) {
	type rspEvent struct {
		Id        string
		Name      string
		StartTime string
		EndTime   string
		MonitorId string
	}

	type rspEventItem struct {
		Event *rspEvent
	}

	type rspPagination struct {
		NextPage bool
	}

	type rsp struct {
		Events     []*rspEventItem
		Pagination *rspPagination
	}

	var result []Event
	page := 1
	for {
		url := fmt.Sprintf("events/index/StartTime >=:%s.json?page=%d", minStart.Format(timeFormat), page)

		var r rsp
		if err := c.doGet(ctx, url, &r); err != nil {
			return nil, fmt.Errorf("Request failed: %v", err)
		}

		// Fetch monitors and index them by ID so we can quickly reference them
		// when filling in the returned events.
		mon, err := c.Monitors(ctx)
		if err != nil {
			return nil, fmt.Errorf("Failed to get monitors: %v", err)
		}
		monitors := make(map[string]*Monitor)
		for _, m := range mon {
			m := m
			monitors[m.Id] = &m
		}

		for _, ei := range r.Events {
			if ei.Event != nil {
				// Some events are not yet complete and thus do not have
				// end times. Ignore those instead of making up data for
				// them.
				if ei.Event.EndTime == "" {
					continue
				}

				start, err := time.Parse(timeFormat, ei.Event.StartTime)
				if err != nil {
					log.Errorf("Failed to parse start time %s; skipping event: %v", ei.Event.StartTime, err)
					continue
				}

				end, err := time.Parse(timeFormat, ei.Event.EndTime)
				if err != nil {
					log.Errorf("Failed to parse end time %s; skipping event: %v", ei.Event.EndTime, err)
					continue
				}

				monitor, present := monitors[ei.Event.MonitorId]
				if !present {
					log.Errorf("Failed to find monitor ID %s; skipping event", ei.Event.MonitorId)
					continue
				}

				result = append(result, Event{
					Id:      ei.Event.Id,
					Name:    ei.Event.Name,
					Start:   start.UTC(),
					End:     end.UTC(),
					Monitor: monitor,
				})
			}
		}

		if r.Pagination != nil && r.Pagination.NextPage {
			// More results to read
			page++
			continue
		}

		// All done.
		break
	}
	return result, nil
}
