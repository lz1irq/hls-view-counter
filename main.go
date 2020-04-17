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
var httpBind string

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
		&httpBind, "http-bind", "",
		"Address and port to bind HTTP JSON export to",
	)

	flag.Parse()
}

func main() {
	fmt.Printf("Using interval=%s, logFile=%s\n", interval, logFile)
	counter := newViewCounter(logFile, interval)

	if httpBind != "" {
		httpExporter := HttpViewCountExporter{}
		go httpExporter.export(httpBind)
		counter.addExporter(&httpExporter)
	}

	counter.countViews()
}
