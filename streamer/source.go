package streamer

import (
	"errors"
	"sync"
	"wssAPI"
)

type streamSource struct {
	bProducer bool
	mutexSink sync.RWMutex
	sinks     map[string]*streamSink
}

func (this *streamSource) Init(msg *wssAPI.Msg) (err error) {
	this.sinks = make(map[string]*streamSink)
	return
}

func (this *streamSource) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSource) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSource) GetType() string {
	return streamTypeSource
}

func (this *streamSource) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *streamSource) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSource) HasProducer() bool {
	return this.bProducer
}

func (this *streamSource) SetProducer(status bool) (remove bool) {
	if status == this.bProducer {
		return
	}
	this.bProducer = status
	if this.bProducer == false {
		//notify sinks stop
		if 0 == len(this.sinks) {
			return true
		}
		this.mutexSink.RLock()
		defer this.mutexSink.RUnlock()
		for _, v := range this.sinks {
			v.Stop(nil)
		}
		return
	} else {
		//notify sinks start
		defer this.mutexSink.RUnlock()
		for _, v := range this.sinks {
			v.Start(nil)
		}
		return
	}
}

func (this *streamSource) AddSink(id string, sinker wssAPI.Obj) (err error) {
	this.mutexSink.Lock()
	defer this.mutexSink.Unlock()
	_, exist := this.sinks[id]
	if true == exist {
		return errors.New("sink " + id + " exist")
	}
	sink := &streamSink{}
	msg := &wssAPI.Msg{}
	msg.Param1 = id
	msg.Param2 = sinker
	err = sink.Init(msg)
	if err != nil {
		return
	}
	this.sinks[id] = sink
	err = sink.Start(nil)
	return
}

func (this *streamSource) DelSink(id string) (err error, removeSrc bool) {
	this.mutexSink.Lock()
	defer this.mutexSink.Unlock()
	sink, exist := this.sinks[id]
	if false == exist {
		return errors.New("sink " + id + " not found"), false
	}
	sink.Stop(nil)
	delete(this.sinks, id)
	//check if delete source
	if 0 == len(this.sinks) && this.bProducer == false {
		removeSrc = true
	}
	return
}
