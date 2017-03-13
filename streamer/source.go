package streamer

import (
	"wssAPI"
)

type streamSource struct {
	bProducer bool
}

func (this *streamSource) Init(msg *wssAPI.Msg) (err error) {
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

func (this *streamSource) SetProducer(status bool) {
	if status == this.bProducer {
		return
	}
	this.bProducer = status
	if this.bProducer == true {
		//notify sinks stop
	} else {
		//no need to start
		return
	}
}
