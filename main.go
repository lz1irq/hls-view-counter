package main

import (
	"flag"
	"fmt"
	"regexp"
	"time"
)

var streamNameRegex = regexp.MustCompile(`\/hls\/(?P<streamName>.*)-\d+\.ts`)

var interval time.Duration
var logFile string
var exportHTTP string
var exportCollectd string

func init() {
	flag.DurationVar(
		&interval, "interval", time.Duration(10*time.Second),
		"Interval between statistics output",
	)

	flag.StringVar(
		&logFile, "logfile", "/var/log/nginx/access.log",
		"Path to Nginx access log file",
	)

	flag.StringVar(
		&exportHTTP, "export.http", "",
		"Address and port to bind HTTP JSON export to (e.g. '127.0.0.1:9966')",
	)

	flag.StringVar(
		&exportCollectd, "export.collectd", "",
		"Collectd Unix socket path to write statistics to (e.g. '/var/run/collectd-unixsock')",
	)

	flag.Parse()
}

func main() {
	fmt.Printf("Using interval=%s, logFile=%s\n", interval, logFile)
	counter := newViewCounter(logFile, interval)

	if exportHTTP != "" {
		httpExporter := HttpViewCountExporter{}
		go httpExporter.export(exportHTTP)
		counter.addExporter(&httpExporter)
	}

	if exportCollectd != "" {
		collectdExporter := CollectdExporter{}
		go collectdExporter.export(exportCollectd)
		counter.addExporter(&collectdExporter)
	}

	counter.countViews()
}
