package webSocketService

import (
	"encoding/json"
	"errors"
	"logger"
	"wssAPI"
)

type WebSocketService struct {
}

type WebSocketConfig struct {
	Port     int    `json:"Port"`
	PlayPath string `json:"PlayPath"`
}

var service *WebSocketService
var serviceConfig WebSocketConfig

func (this *WebSocketService) Init(msg *wssAPI.Msg) (err error) {

	if msg == nil || msg.Param1 == nil {
		logger.LOGE("init Websocket service failed")
		return errors.New("invalid param")
	}

	fileName := msg.Param1.(string)
	err = this.loadConfigFile(fileName)
	if err != nil {
		logger.LOGE(err.Error())
		return errors.New("load websocket config failed")
	}
	service = this
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

func (this *WebSocketService) loadConfigFile(fileName string) (err error) {
	data, err := wssAPI.ReadFileAll(fileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &serviceConfig)
	if err != nil {
		return
	}

	return
}
