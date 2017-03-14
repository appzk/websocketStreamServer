package main

import (
	"logger"
	"svrbus"
)

func main() {
	initLogger()
	startServers()
}

func initLogger() {
	logger.SetFlags(logger.LOG_SHORT_FILE | logger.LOG_TIME)
	logger.SetLogLevel(logger.LOG_LEVEL_INFO)
}

func startServers() {
	bus := &svrBus.SvrBus{}
	err := bus.Init(nil)
	if err != nil {
		logger.LOGF(err.Error())
	}
	err = bus.Start(nil)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	ch := make(chan int)
	<-ch
}
