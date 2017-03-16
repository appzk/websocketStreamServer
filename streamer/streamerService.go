package streamer

import (
	"errors"
	"logger"
	"sync"
	"wssAPI"
)

const (
	streamTypeService = "streamService"
	streamTypeSource  = "streamSource"
	streamTypeSink    = "streamSink"
)

type StreamerService struct {
	mutexSources sync.RWMutex
	sources      map[string]*streamSource
}

var service *StreamerService

func (this *StreamerService) Init(msg *wssAPI.Msg) (err error) {
	this.sources = make(map[string]*streamSource)
	service = this
	return
}

func (this *StreamerService) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *StreamerService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *StreamerService) GetType() string {
	return streamTypeService
}

func (this *StreamerService) HandleTask(task *wssAPI.Task) (err error) {
	switch task.Type {
	case wssAPI.TASK_AddSource:
	case wssAPI.TASK_DelSource:
	case wssAPI.TASK_AddSink:
	case wssAPI.TASK_DelSink:
	default:
		logger.LOGW(task.Type + " should not handle here")
	}
	return
}

func (this *StreamerService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

//src control sink
//add source:not start src,start sinks
//del source:not stop src,stop sinks
func Addsource(path string) (src wssAPI.Obj, err error) {
	if service == nil {
		logger.LOGE("streamer service null")
		err = errors.New("streamer invalid")
		return
	}

	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	logger.LOGT("add source:" + path)
	oldSrc, exist := service.sources[path]
	if exist == false {
		oldSrc = &streamSource{}
		oldSrc.SetProducer(true)
		service.sources[path] = oldSrc
		src = oldSrc
		return
	} else {
		if oldSrc.HasProducer() {
			err = errors.New("bad name")
			return
		} else {
			logger.LOGT("source:" + path + " is idle")
			oldSrc.SetProducer(true)
			src = oldSrc
			return
		}
	}
	return
}

func DelSource(path string) (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	logger.LOGT("del source:" + path)
	oldSrc, exist := service.sources[path]
	if exist == false {
		return errors.New(path + " not found")
	} else {
		remove := oldSrc.SetProducer(false)
		if remove == true {
			delete(service.sources, path)
		}
		return
	}
	return
}

//add sink:auto start sink by src
//del sink:not stop sink,stop by sink itself
func AddSink(path, sinkId string, sinker wssAPI.Obj) (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	src, exist := service.sources[path]
	if false == exist {
		err = errors.New("source not found in add sink")
		return
	} else {
		return src.AddSink(sinkId, sinker)
	}
	return
}

func DelSink(path, sinkId string) (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	src, exist := service.sources[path]
	if false == exist {
		return errors.New("source not found in del sink")
	} else {
		removeSrc := false
		err, removeSrc = src.DelSink(sinkId)
		if err != nil {
			logger.LOGE(err.Error())
			return errors.New("delete sink" + sinkId + " failed")
		}
		if true == removeSrc {
			delete(service.sources, path)
		}
	}
	return
}
