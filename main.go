package main

import (
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/nxadm/tail"
)

var streamNameRegex = regexp.MustCompile(`\/hls\/(?P<streamName>.*)-\d+\.ts`)
var streamViewers = map[string]map[string]struct{}{}

var interval time.Duration
var logFile string

func init() {
	flag.DurationVar(
		&interval, "interval", time.Duration(10*time.Second),
		"Interval between statistics output",
	)

	flag.StringVar(
		&logFile, "logfile", "/var/log/nginx/access.log",
		"Path to Nginx access log file",
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

	matches := streamNameRegex.FindAllStringSubmatch(url, 1)
	if len(matches) == 0 {
		return
	}
	streamName := matches[0][1]

	if _, ok := streamViewers[streamName]; !ok {
		streamViewers[streamName] = map[string]struct{}{}
	}
	streamViewers[streamName][ip] = struct{}{}
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
			for streamName, viewers := range streamViewers {
				fmt.Printf("stream=%s, viewers=%d\n", streamName, len(viewers))
			}
			streamViewers = map[string]map[string]struct{}{}
		}
	}
}

func main() {
	fmt.Printf("Using interval=%s, logFile=%s\n", interval, logFile)
	countViews(logFile)
}
