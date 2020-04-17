package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nxadm/tail"
)

var streamNameRegex = regexp.MustCompile(`\/hls\/(?P<streamName>.*)-\d+\.ts`)
var streamViewers = map[string]map[string]struct{}{}
var exportViewers = map[string]int{}

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

func readLines(logFile string, out chan string) {
	t, err := tail.TailFile(
		logFile,
		tail.Config{
			Follow: true,
			Location: &tail.SeekInfo{
				Whence: io.SeekEnd,
			},
		},
	)
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}
	for line := range t.Lines {
		out <- line.Text
	}
}

func processLine(line string) {
	parts := strings.Split(line, " ")
	ip := parts[0]
	url := parts[6]

	match := streamNameRegex.FindStringSubmatch(url)
	if len(match) == 0 {
		return
	}
	streamName := match[1]

	streamViewersPerStream, ok := streamViewers[streamName]
	if !ok {
		streamViewersPerStream = map[string]struct{}{}
		streamViewers[streamName] = streamViewersPerStream
	}
	streamViewersPerStream[ip] = struct{}{}
}

func countViews(logFile string) {
	lineChan := make(chan string, 1000)
	go readLines(logFile, lineChan)

	ticker := time.NewTicker(interval)
	for {
		select {
		case line := <-lineChan:
			processLine(line)
		case <-ticker.C:
			exportViewers = map[string]int{}
			for streamName, viewers := range streamViewers {
				fmt.Printf("stream=%s, viewers=%d\n", streamName, len(viewers))
				exportViewers[streamName] = len(viewers)
			}
			streamViewers = map[string]map[string]struct{}{}
		}
	}
}

type httpViewsData struct {
	Streams map[string]int `json:"streams"`
}

func runHTTPExport() {
	fmt.Printf("Binding HTTP export on %s\n", httpBind)
	r := mux.NewRouter()
	r.HandleFunc(
		"/views",
		func(w http.ResponseWriter, r *http.Request) {
			data := httpViewsData{exportViewers}
			json.NewEncoder(w).Encode(data)
		},
	)
	http.ListenAndServe(httpBind, r)
}

func main() {
	fmt.Printf("Using interval=%s, logFile=%s\n", interval, logFile)
	if httpBind != "" {
		go runHTTPExport()
	}
	countViews(logFile)
}
