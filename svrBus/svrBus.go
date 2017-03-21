package svrBus

import (
	"RTMPService"
	"encoding/json"
	"errors"
	"logger"
	"os"
	"runtime"
	"streamer"
	"strings"
	"sync"
	"time"
	"webSocketService"
	"wssAPI"
)

type busConfig struct {
	RTMPConfigName      string `json:"RTMP"`
	WebSocketConfigName string `json:"WebSocket"`
	LogPath             string `json:"LogPath"`
}

type SvrBus struct {
	mutexServices sync.RWMutex
	services      map[string]wssAPI.Obj
}

func (this *SvrBus) Init(msg *wssAPI.Msg) (err error) {
	this.services = make(map[string]wssAPI.Obj)
	err = this.loadConfig()
	if err != nil {
		logger.LOGE("svr bus load config failed")
		return
	}
	return
}

func (this *SvrBus) loadConfig() (err error) {
	var configName string
	if len(os.Args) > 1 {
		configName = os.Args[1]
	} else {
		logger.LOGW("use default :config.json")
		configName = "config.json"
	}
	data, err := wssAPI.ReadFileAll(configName)
	if err != nil {
		logger.LOGE("load config file failed:" + err.Error())
		return
	}
	cfg := &busConfig{}
	err = json.Unmarshal(data, cfg)

	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	if len(cfg.LogPath) > 0 {
		this.createLogFile(cfg.LogPath)
	}

	if true {
		livingSvr := &streamer.StreamerService{}
		livingSvr.Init(nil)
		this.mutexServices.Lock()
		this.services[livingSvr.GetType()] = livingSvr
		this.mutexServices.Unlock()
	}

	if len(cfg.RTMPConfigName) > 0 {
		rtmpSvr := &RTMPService.RTMPService{}
		msg := &wssAPI.Msg{}
		msg.Param1 = cfg.RTMPConfigName
		rtmpSvr.Init(msg)
		this.mutexServices.Lock()
		this.services[rtmpSvr.GetType()] = rtmpSvr
		this.mutexServices.Unlock()
	}

	if len(cfg.WebSocketConfigName) > 0 {
		webSocketSvr := &webSocketService.WebSocketService{}
		msg := &wssAPI.Msg{}
		msg.Param1 = cfg.WebSocketConfigName
		webSocketSvr.Init(msg)
		this.mutexServices.Lock()
		this.services[webSocketSvr.GetType()] = webSocketSvr
		this.mutexServices.Unlock()
	}

	return
}

func (this *SvrBus) createLogFile(logPath string) {
	if strings.HasSuffix(logPath, "/") {
		logPath = strings.TrimSuffix(logPath, "/")
	}
	dir := logPath + time.Now().Format("/2006/01/02/")
	bResult, _ := wssAPI.CheckDirectory(dir)

	if false == bResult {
		_, err := wssAPI.CreateDirectory(dir)
		if err != nil {
			logger.LOGE("create log file failed:", err.Error())
			return
		}
	}
	fullName := dir + time.Now().Format("2006-01-02_15.04") + ".log"
	fp, err := os.OpenFile(fullName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, os.ModePerm)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	logger.SetOutput(fp)
	//avoid one log file too big
	go func() {
		logFileTick := time.Tick(time.Hour * 72)
		for {
			select {
			case <-logFileTick:
				fullName := dir + time.Now().Format("2006-01-02_15:04") + ".log"
				newLogFile, _ := os.OpenFile(fullName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, os.ModePerm)
				if newLogFile != nil {
					logger.SetOutput(newLogFile)
					fp.Close()
					fp = newLogFile
				}
			}
		}
	}()
}

func (this *SvrBus) Start(msg *wssAPI.Obj) (err error) {
	if false {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	this.mutexServices.RLock()
	defer this.mutexServices.RUnlock()
	for k, v := range this.services {
		err = v.Start(nil)
		if err != nil {
			logger.LOGE("start " + k + " failed:" + err.Error())
			continue
		}
		logger.LOGI("start " + k + " successed")
	}
	return
}

func (this *SvrBus) Stop(msg *wssAPI.Obj) (err error) {
	this.mutexServices.RLock()
	defer this.mutexServices.RUnlock()
	for _, v := range this.services {
		err = v.Stop(nil)
	}
	return
}

func (this *SvrBus) GetType() string {
	return wssAPI.OBJ_ServerBus
}

func (this *SvrBus) HandleTask(task *wssAPI.Task) (err error) {
	this.mutexServices.RLock()
	defer this.mutexServices.RUnlock()
	handler, exist := this.services[task.Reciver]
	if exist == false {
		return errors.New("invalid task")
	}
	return handler.HandleTask(task)
}

func (this *SvrBus) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return nil
}
