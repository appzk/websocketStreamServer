package streamer

import (
	"wssAPI"
)

type streamSink struct {
}

func (this *streamSink) Init(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSink) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSink) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSink) GetType() string {
	return streamTypeSink
}

func (this *streamSink) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *streamSink) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}
