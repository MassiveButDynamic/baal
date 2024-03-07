package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var upgrader = websocket.Upgrader{}

var (
	accessCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "apache_access_count",
	}, []string{"site", "method", "status"})

	accessBytesSent = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "apache_access_bytes_sent",
	}, []string{"site", "method", "status"})

	accessDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "apache_access_duration_us",
	}, []string{"site", "method", "status"})
)

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read failed: ", err)
				break
			}
			input := string(msg)
			parseLogLine(input)
			// log.Print(input)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "baal-daemon")
	})

	http.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":23873", nil)
}

// Example:
// 2024/03/07 15:28:06 46.142.235.75|-|-|[07/Mar/2024:15:28:06 +0100]|"GET /assets/favicon/favicon-16x16.png HTTP/2.0"|200|1570|1552
//
// Format: %h|%l|%u|%t|\"%r\"|%>s|%O|%{us}T
// h: Remote hostname, l: Remote logname, u: Remote user, t: Time, r: First line of request, s: Status, O: Bytes sent, T: time taken for the request in microseconds
func parseLogLine(line string) {
	segs := strings.Split(line, "|")

	if len(segs) != 9 {
		fmt.Println("Error: log line does not have 9 segments")
		fmt.Println(line)
		return
	}

	requestLineSegs := strings.Split(segs[4], " ")
	method := strings.ReplaceAll(requestLineSegs[0], "\"", "")
	status := segs[5]
	bytesSent, _ := strconv.Atoi(segs[6])
	timeUs, _ := strconv.Atoi(segs[7])
	site := segs[8]

	accessCounter.WithLabelValues(site, method, status).Inc()
	accessBytesSent.WithLabelValues(site, method, status).Observe(float64(bytesSent))
	accessDuration.WithLabelValues(site, method, status).Observe(float64(timeUs))

	fmt.Printf("Site: %s, Method: %s, Status: %s, Time: %dus\n", site, method, status, timeUs)
}
