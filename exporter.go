package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type ViewCountExporter interface {
	updateViewCount(newViews map[string]int, updatedAt int64)
	export(config string)
}

// HTTP exporter

type httpViewsData struct {
	Streams   map[string]int `json:"streams"`
	UpdatedAt int64          `json:"updatedAt"`
}

type HttpViewCountExporter struct {
	views []byte
}

func (h *HttpViewCountExporter) updateViewCount(newViews map[string]int, updatedAt int64) {
	httpViews := httpViewsData{
		Streams:   newViews,
		UpdatedAt: updatedAt,
	}
	h.views, _ = json.Marshal(httpViews)
}

func (h *HttpViewCountExporter) export(config string) {
	fmt.Printf("Binding HTTP export on %s\n", config)
	r := mux.NewRouter()
	r.HandleFunc(
		"/views",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write(h.views)
		},
	)
	http.ListenAndServe(config, r)
}
