# hls-view-counter

Small Go program to count the approximate number of viewers of HLS streams
running over [nginx-rtmp-module](https://github.com/arut/nginx-rtmp-module)
using the access log.

## Usage

Requires [Go](https://golang.org/dl/).

```
git clone https://github.com/lz1irq/hls-view-counter.git
cd hls-view-counter
go build
./hls-view-counter -interval 2s -logfile /var/log/nginx/access.log.1
```
