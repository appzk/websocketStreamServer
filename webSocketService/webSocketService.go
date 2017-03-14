package webSocketService

import (
	"wssAPI"
)

type WebSocketService struct {
}

func (this *WebSocketService) Init(msg *wssAPI.Msg) (err error) {
	return
}

func (this *WebSocketService) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *WebSocketService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *WebSocketService) GetType() string {
	return wssAPI.OBJ_WebSocketServer
}

func (this *WebSocketService) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *WebSocketService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}
