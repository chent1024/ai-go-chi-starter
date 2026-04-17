package runtime

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	startedAt time.Time
	build     BuildInfo

	inFlight   atomic.Int64
	mu         sync.RWMutex
	requests   map[requestMetricKey]*requestMetricValue
	lateWrites map[string]uint64
}

type requestMetricKey struct {
	Route  string
	Method string
	Status int
}

type requestMetricValue struct {
	Count        uint64
	LatencyMsSum int64
	LatencyMsMax int64
}

func NewMetrics(build BuildInfo) *Metrics {
	return &Metrics{
		startedAt:  time.Now().UTC(),
		build:      build,
		requests:   make(map[requestMetricKey]*requestMetricValue),
		lateWrites: make(map[string]uint64),
	}
}

func (m *Metrics) ObserveHTTPRequest(route, method string, status int, latency time.Duration) {
	if m == nil {
		return
	}
	key := requestMetricKey{
		Route:  defaultMetricRoute(route),
		Method: strings.ToUpper(strings.TrimSpace(method)),
		Status: status,
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	value := m.requests[key]
	if value == nil {
		value = &requestMetricValue{}
		m.requests[key] = value
	}
	latencyMs := latency.Milliseconds()
	value.Count++
	value.LatencyMsSum += latencyMs
	if latencyMs > value.LatencyMsMax {
		value.LatencyMsMax = latencyMs
	}
}

func (m *Metrics) IncInFlight() {
	if m == nil {
		return
	}
	m.inFlight.Add(1)
}

func (m *Metrics) DecInFlight() {
	if m == nil {
		return
	}
	m.inFlight.Add(-1)
}

func (m *Metrics) ObserveTimeoutLateWrite(route string, count int) {
	if m == nil || count <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	route = defaultMetricRoute(route)
	m.lateWrites[route] += uint64(count)
}

func (m *Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if m == nil {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = fmt.Fprintf(
		w,
		"app_build_info{service=%q,version=%q,commit=%q,build_time=%q} 1\n",
		escapeMetricLabel(m.build.Service),
		escapeMetricLabel(m.build.Version),
		escapeMetricLabel(m.build.Commit),
		escapeMetricLabel(m.build.BuildTime),
	)
	_, _ = fmt.Fprintf(w, "process_start_time_seconds %d\n", m.startedAt.Unix())
	_, _ = fmt.Fprintf(w, "process_uptime_seconds %d\n", int64(time.Since(m.startedAt).Seconds()))
	_, _ = fmt.Fprintf(w, "http_requests_in_flight %d\n", m.inFlight.Load())

	requests, lateWrites := m.snapshot()

	for _, series := range requests {
		_, _ = fmt.Fprintf(
			w,
			"http_requests_total{route=%q,method=%q,status=%q} %d\n",
			escapeMetricLabel(series.Route),
			escapeMetricLabel(series.Method),
			escapeMetricLabel(strconv.Itoa(series.Status)),
			series.Count,
		)
		_, _ = fmt.Fprintf(
			w,
			"http_request_duration_ms_sum{route=%q,method=%q,status=%q} %d\n",
			escapeMetricLabel(series.Route),
			escapeMetricLabel(series.Method),
			escapeMetricLabel(strconv.Itoa(series.Status)),
			series.LatencyMsSum,
		)
		_, _ = fmt.Fprintf(
			w,
			"http_request_duration_ms_count{route=%q,method=%q,status=%q} %d\n",
			escapeMetricLabel(series.Route),
			escapeMetricLabel(series.Method),
			escapeMetricLabel(strconv.Itoa(series.Status)),
			series.Count,
		)
		_, _ = fmt.Fprintf(
			w,
			"http_request_duration_ms_max{route=%q,method=%q,status=%q} %d\n",
			escapeMetricLabel(series.Route),
			escapeMetricLabel(series.Method),
			escapeMetricLabel(strconv.Itoa(series.Status)),
			series.LatencyMsMax,
		)
	}

	for _, series := range lateWrites {
		_, _ = fmt.Fprintf(
			w,
			"http_request_timeout_late_write_total{route=%q} %d\n",
			escapeMetricLabel(series.Route),
			series.Count,
		)
	}
}

type metricSeries struct {
	Route        string
	Method       string
	Status       int
	Count        uint64
	LatencyMsSum int64
	LatencyMsMax int64
}

func (m *Metrics) snapshot() ([]metricSeries, []metricSeries) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requests := make([]metricSeries, 0, len(m.requests))
	for key, value := range m.requests {
		requests = append(requests, metricSeries{
			Route:        key.Route,
			Method:       key.Method,
			Status:       key.Status,
			Count:        value.Count,
			LatencyMsSum: value.LatencyMsSum,
			LatencyMsMax: value.LatencyMsMax,
		})
	}
	sort.Slice(requests, func(i, j int) bool {
		if requests[i].Route != requests[j].Route {
			return requests[i].Route < requests[j].Route
		}
		if requests[i].Method != requests[j].Method {
			return requests[i].Method < requests[j].Method
		}
		return requests[i].Status < requests[j].Status
	})

	lateWrites := make([]metricSeries, 0, len(m.lateWrites))
	for route, count := range m.lateWrites {
		lateWrites = append(lateWrites, metricSeries{
			Route: route,
			Count: count,
		})
	}
	sort.Slice(lateWrites, func(i, j int) bool {
		return lateWrites[i].Route < lateWrites[j].Route
	})

	return requests, lateWrites
}

func defaultMetricRoute(route string) string {
	route = strings.TrimSpace(route)
	if route == "" {
		return "unmatched"
	}
	return route
}

func escapeMetricLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}
