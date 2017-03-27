package webSocketService

import (
	"encoding/json"
	"errors"
	"logger"
	"net/http"
	"strconv"
	"strings"
	"wssAPI"

	"github.com/gorilla/websocket"
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

	go func() {
		strPort := ":" + strconv.Itoa(serviceConfig.Port)
		if len(serviceConfig.PlayPath) > 0 {
			if strings.HasPrefix(serviceConfig.PlayPath, "/") == false {
				serviceConfig.PlayPath = "/" + serviceConfig.PlayPath
			}
		}
		http.Handle(serviceConfig.PlayPath, http.StripPrefix(serviceConfig.PlayPath, this))
		err = http.ListenAndServe(strPort, nil)
		if err != nil {
			logger.LOGE("start websocket failed:" + err.Error())
		}
	}()
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

func (this *WebSocketService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	defer func() {
		conn.Close()
	}()
}
