package RTMPService

import (
	"encoding/json"
	"errors"
	"logger"
	"net"
	"strconv"
	"wssAPI"
)

const (
	rtmpTypeHandler = "rtmpHandler"
	livePathDefault = "live"
	timeoutDefault  = 3000
)

type RTMPService struct {
	listener *net.TCPListener
}

type RTMPConfig struct {
	Port       int    `json:"Port"`
	TimeoutSec int    `json:"TimeoutSec"`
	LivePath   string `json:"LivePath"`
}

var service *RTMPService
var serviceConfig RTMPConfig

func (this *RTMPService) Init(msg *wssAPI.Msg) (err error) {
	if nil == msg || nil == msg.Param1 {
		logger.LOGE("invalid param init rtmp server")
		return errors.New("init rtmp service failed")
	}
	fileName := msg.Param1.(string)
	err = this.loadConfigFile(fileName)
	if err != nil {
		logger.LOGE(err.Error())
		return errors.New("init rtmp service failed")
	}
	service = this
	return
}

func (this *RTMPService) Start(msg *wssAPI.Msg) (err error) {
	logger.LOGT("start rtmp service")
	strPort := ":" + strconv.Itoa(serviceConfig.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", strPort)
	if nil != err {
		logger.LOGE(err.Error())
		return
	}
	this.listener, err = net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		logger.LOGE(err.Error())
	}
	go this.rtmpLoop()
	return
}

func (this *RTMPService) Stop(msg *wssAPI.Msg) (err error) {
	this.listener.Close()
	return
}

func (this *RTMPService) GetType() string {
	return wssAPI.OBJ_RTMPServer
}

func (this *RTMPService) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *RTMPService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPService) loadConfigFile(fileName string) (err error) {
	data, err := wssAPI.ReadFileAll(fileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &serviceConfig)
	if err != nil {
		return
	}

	if serviceConfig.TimeoutSec == 0 {
		serviceConfig.TimeoutSec = timeoutDefault
	}

	if len(serviceConfig.LivePath) == 0 {
		serviceConfig.LivePath = livePathDefault
	}
	strPort := ""
	if strPort != "1935" {
		strPort = strconv.Itoa(serviceConfig.Port)
	}
	logger.LOGI("rtmp://address:" + strPort + "/" + serviceConfig.LivePath + "/streamName")
	logger.LOGI("rtmp timeout: " + strconv.Itoa(serviceConfig.TimeoutSec) + " s")
	return
}

func (this *RTMPService) rtmpLoop() {
	for {
		conn, err := this.listener.Accept()
		if err != nil {
			logger.LOGW(err.Error())
			continue
		}
		go this.handleConnect(conn)
	}
}

func (this *RTMPService) handleConnect(conn net.Conn) {
	var err error
	defer conn.Close()
	err = rtmpHandleshake(conn)
	if err != nil {
		logger.LOGE("rtmp handle shake failed")
		return
	}

}
