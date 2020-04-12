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
var streamViewers = map[string]map[string]byte{}

var interval uint
var logFile string

func init() {
	flag.UintVar(
		&interval, "interval", 10,
		"Interval in seconds between statistics output",
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
		streamViewers[streamName] = map[string]byte{}
	}
	streamViewers[streamName][ip] = 0
}

func countViews(logFile string) {
	lineChan := make(chan string, 1000)
	go readLines(logFile, lineChan)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case line := <-lineChan:
			processLine(line)
		case <-ticker.C:
			for streamName, viewers := range streamViewers {
				fmt.Printf("stream=%s, viewers=%d\n", streamName, len(viewers))
			}
			streamViewers = map[string]map[string]byte{}
		}
	}
}

func main() {
	fmt.Printf("Using logFile=%s\n", logFile)
	countViews(logFile)
}
