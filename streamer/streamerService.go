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

func Addsource(path string, publisher wssAPI.Obj) (src wssAPI.Obj, err error) {
	if service == nil {
		err = errors.New("streamer invalid")
		return
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
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
	oldSrc, exist := service.sources[path]
	if exist == false {
		return errors.New(path + " not found")
	} else {
		oldSrc.SetProducer(false)
		return
	}
	return
}

func AddSink() (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	return
}

func DelSink() (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	return
}
