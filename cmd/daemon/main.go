package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

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

	requestLineSegs := strings.Split(segs[5], " ")
	method := requestLineSegs[0]
	status, _ := strconv.Atoi(segs[6])
	timeUs, _ := strconv.Atoi(segs[7])
	site := segs[8]

	fmt.Printf("Site: %s, Method: %s, Status: %d, Time: %dus\n", site, method, status, timeUs)
}
