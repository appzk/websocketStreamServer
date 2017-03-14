package RTMPService

import (
	"wssAPI"
)

type RTMPHandler struct {
}

func (this *RTMPHandler) Init(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPHandler) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPHandler) GetType() string {
	return rtmpTypeHandler
}

func (this *RTMPHandler) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *RTMPHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}
