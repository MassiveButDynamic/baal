package main

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:23873", "baal daemon address")
var site = flag.String("site", "", "apache sitename")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	msg := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg <- scanner.Text() + "|" + *site
		}
	}()

loop:
	for {
		select {
		case <-sigs:
			break loop
		case s := <-msg:
			err := c.WriteMessage(websocket.TextMessage, []byte(s))
			if err != nil {
				break loop
			}
		}
	}
}
