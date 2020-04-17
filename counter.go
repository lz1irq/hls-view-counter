package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nxadm/tail"
)

type viewCounter struct {
	logFile         string
	interval        time.Duration
	exporters       []ViewCountExporter
	streamViewers   map[string]map[string]struct{}
	exportViews     map[string]int
	exportUpdatedAt int64
}

func newViewCounter(logFile string, interval time.Duration) *viewCounter {
	return &viewCounter{
		logFile:       logFile,
		interval:      interval,
		streamViewers: map[string]map[string]struct{}{},
		exportViews:   map[string]int{},
	}
}

func (v *viewCounter) addExporter(exporter ViewCountExporter) {
	v.exporters = append(v.exporters, exporter)
}

func (v *viewCounter) updateExporters() {
	for _, exporter := range v.exporters {
		exporter.updateViewCount(v.exportViews, v.exportUpdatedAt)
	}
}

func (v *viewCounter) readLines(out chan string) {
	t, err := tail.TailFile(
		v.logFile,
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

func (v *viewCounter) processLine(line string) {
	parts := strings.Split(line, " ")
	ip := parts[0]
	url := parts[6]

	match := streamNameRegex.FindStringSubmatch(url)
	if len(match) == 0 {
		return
	}
	streamName := match[1]

	streamViewersPerStream, ok := v.streamViewers[streamName]
	if !ok {
		streamViewersPerStream = map[string]struct{}{}
		v.streamViewers[streamName] = streamViewersPerStream
	}
	streamViewersPerStream[ip] = struct{}{}
}

func (v *viewCounter) countViews() {
	lineChan := make(chan string, 1000)
	go v.readLines(lineChan)

	ticker := time.NewTicker(interval)
	for {
		select {
		case line := <-lineChan:
			v.processLine(line)
		case <-ticker.C:
			v.exportViews = map[string]int{}
			for streamName, viewers := range v.streamViewers {
				fmt.Printf("stream=%s, viewers=%d\n", streamName, len(viewers))
				v.exportViews[streamName] = len(viewers)
			}
			v.exportUpdatedAt = time.Now().Unix()
			v.updateExporters()

			v.streamViewers = map[string]map[string]struct{}{}
		}
	}
}
