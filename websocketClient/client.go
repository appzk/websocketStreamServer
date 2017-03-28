package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	cli := &websocket.Dialer{}
	req := http.Header{}
	conn, res, err := cli.Dial("ws://127.0.0.1:8080/play", req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer conn.Close()
	conn.WriteMessage(websocket.TextMessage, []byte("a ha  sdfsfadfasd"))
	_, data, err := conn.ReadMessage()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(string(data))
	log.Println(res.Close)
}
