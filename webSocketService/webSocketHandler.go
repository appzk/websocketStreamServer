package webSocketService

import (
	"wssAPI"
)

const (
	wsHandler = "websocketHandler"
)

type websocketHandler struct {
}

func (this *websocketHandler) Init(msg *wssAPI.Msg) (err error) {
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

func (this *websocketHandler) processWSControl(data []byte) (err error) {
	return
}
