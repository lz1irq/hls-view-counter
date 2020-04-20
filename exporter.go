package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"

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

// Collectd exporter

const collectdSocketType = "unix"
const collectdPluginName = "nginx_rtmp"
const collectdDataType = "gauge"
const collectdValueName = "hls_viewers"

type CollectdExporter struct {
	hostname string
	socket   net.Conn
}

func (c *CollectdExporter) getHostname() string {
	cmd := exec.Command("/bin/hostname", "-f")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Errorf(err.Error())
	}
	fqdn := out.String()
	fqdn = fqdn[:len(fqdn)-1] // removing EOL
	return fqdn
}

func (c *CollectdExporter) export(sockAddr string) {
	var err error
	c.hostname = c.getHostname()
	c.socket, err = net.Dial(collectdSocketType, sockAddr)
	if err != nil {
		fmt.Errorf("%s\n", err.Error())
	}
}

func (c *CollectdExporter) updateViewCount(newViews map[string]int, updatedAt int64) {
	for streamName, viewCount := range newViews {
		statLine := fmt.Sprintf(
			"PUTVAL %s/%s-%s/%s-%s %d:%d",
			c.hostname,
			collectdPluginName, streamName,
			collectdDataType, collectdValueName,
			updatedAt, viewCount,
		)
		fmt.Printf("%s\n", statLine)
		c.socket.Write([]byte(statLine))

	}
}
