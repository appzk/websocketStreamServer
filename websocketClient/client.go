package main

import (
	"encoding/json"
	"logger"
	"net/http"
	"webSocketService"

	"github.com/gorilla/websocket"
)

func main() {
	logger.SetFlags(logger.LOG_SHORT_FILE)
	cli := &websocket.Dialer{}
	req := http.Header{}
	conn, _, err := cli.Dial("ws://127.0.0.1:8080/live", req)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	Play(conn)
	defer conn.Close()

}

func Play(conn *websocket.Conn) {
	//connect
	stConnect := &webSocketService.WsConnect{}
	stConnect.App = "live"
	stConnect.ID = 0
	stConnect.TcUrl = "ws://127.0.0.1:8080/live"
	dataSend, _ := json.Marshal(stConnect)
	err := webSocketService.SendWsControl(conn, webSocketService.WS_ctrl_connect, dataSend)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	err = readResult(conn)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	//play
	stPlay := &webSocketService.WsPlay{}
	stPlay.ID = 0
	stPlay.StreamName = "test"
	dataSend, _ = json.Marshal(stPlay)
	err = webSocketService.SendWsControl(conn, webSocketService.WS_ctrl_play, dataSend)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	err = readResult(conn)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	err = readResult(conn)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	//read packet
	//stopPlay
	stStop := &webSocketService.WsStopPlay{}
	stStop.ID = 0
	dataSend, _ = json.Marshal(stStop)
	//err = webSocketService.SendWsControl(conn, webSocketService.WS_ctrl_stopPlay, dataSend)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	err = readResult(conn)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	for {
		readResult(conn)
	}
	//end
	return
}

func readResult(conn *websocket.Conn) (err error) {
	msgType, data, err := conn.ReadMessage()
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	if msgType == websocket.BinaryMessage {
		pktType := data[0]
		switch pktType {
		case webSocketService.WS_pkt_audio:
			logger.LOGT("audio")
		case webSocketService.WS_pkt_video:
			logger.LOGT("video")
		case webSocketService.WS_pkt_control:
			switch data[1] {
			case webSocketService.WS_ctrl_result:
				stResult := &webSocketService.WsResult{}
				json.Unmarshal(data[2:], stResult)
				logger.LOGT(*stResult)
			case webSocketService.WS_ctrl_streamBegin:
				logger.LOGT("stream begin")
			case webSocketService.WS_ctrl_streamEnd:
				logger.LOGT("stream end")
			}

		}
	}
	return
}
