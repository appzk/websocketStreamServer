package webSocketService

import (
	"encoding/json"
	"errors"
	"fmt"
	"logger"
	"wssAPI"

	"github.com/gorilla/websocket"
)

const (
	wsHandler = "websocketHandler"
)

type websocketHandler struct {
	conn *websocket.Conn
	app  string
}

func (this *websocketHandler) Init(msg *wssAPI.Msg) (err error) {
	this.conn = msg.Param1.(*websocket.Conn)
	return
}

func (this *websocketHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) GetType() string {
	return wsHandler
}

func (this *websocketHandler) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *websocketHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) processWSMessage(data []byte) (err error) {
	if len(data) < 1 {
		logger.LOGE("invalid message")
		return errors.New("process stream message failed")
	}
	msgType := int(data[0])
	switch msgType {
	case WS_pkt_audio:
	case WS_pkt_video:
	case WS_pkt_control:
		ctrlType := int(data[1])
		switch ctrlType {
		case WS_ctrl_connect:
			stConnect := &WsConnect{}
			err = json.Unmarshal(data[1:], stConnect)
			if err != nil {
				return
			}
			this.app = stConnect.App
			return this.result(stConnect.ID, WS_status_ok, "")
		case WS_ctrl_result:
		case WS_ctrl_play:
		case WS_ctrl_play2:
		case WS_ctrl_pause:
		case WS_ctrl_resume:
		case ws_ctrl_close:
		case WS_ctrl_publish:
		case WS_ctrl_onMetaData:
		case WS_ctrl_unPublish:
		case WS_ctrl_stopPlay:
		}

	default:
		err = errors.New(fmt.Sprintf("msg type %d not supported", msgType))
		return
	}
	return
}

func (this *websocketHandler) result(id, status int, desc string) (err error) {
	stResult := &WsResult{id, status, desc}
	jsondata, err := json.Marshal(stResult)
	if err != nil {
		logger.LOGE("marshal json failed")
		return
	}
	dataSend := make([]byte, len(jsondata)+2)
	dataSend[0] = WS_pkt_control
	dataSend[1] = WS_ctrl_result
	copy(dataSend[2:], jsondata)
	err = this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
	return
}
