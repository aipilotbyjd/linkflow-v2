package metrics

import (
	"encoding/json"
	"net/http"
)

type Exporter struct {
	collector *Collector
}

func NewExporter(collector *Collector) *Exporter {
	return &Exporter{collector: collector}
}

func (e *Exporter) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := e.collector.Snapshot()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	}
}

func (e *Exporter) JSON() ([]byte, error) {
	return json.Marshal(e.collector.Snapshot())
}

func (e *Exporter) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := e.collector.Snapshot()

		status := "healthy"
		code := http.StatusOK

		if !snapshot.IsLeader {
			status = "standby"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    status,
			"is_leader": snapshot.IsLeader,
			"uptime_s":  int64(snapshot.Uptime.Seconds()),
		})
	}
}
