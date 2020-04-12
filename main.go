package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nxadm/tail"
)

var streamNameRegex = regexp.MustCompile(`\/hls\/(?P<streamName>.*)-\d+\.ts`)
var streamViewers = map[string]map[string]byte{}

const interval = 10

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

	ticker := time.NewTicker(interval * time.Second)
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
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s /path/to/access.log", os.Args[0])
		os.Exit(5)
	}
	logFile := os.Args[1]
	fmt.Printf("Using logFile=%s\n", logFile)
	countViews(logFile)
}
