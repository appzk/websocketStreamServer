package streamer

import (
	"errors"
	"logger"
	"wssAPI"
)

type streamSink struct {
	id     string
	sinker wssAPI.Obj
}

func (this *streamSink) Init(msg *wssAPI.Msg) (err error) {
	if nil == msg || msg.Param1 == nil || msg.Param2 == nil {
		return errors.New("invalid init stream sink")
	}
	this.id = msg.Param1.(string)
	this.sinker = msg.Param2.(wssAPI.Obj)
	return
}

func (this *streamSink) Start(msg *wssAPI.Msg) (err error) {
	//notify sinker stream start
	if this.sinker == nil {
		logger.LOGE("sinker no seted")
		return errors.New("no sinker to start")
	}
	msg = &wssAPI.Msg{}
	msg.Type = wssAPI.MSG_STREAM_START
	err = this.sinker.ProcessMessage(msg)
	return
}

func (this *streamSink) Stop(msg *wssAPI.Msg) (err error) {
	//notify sinker stream stop
	if this.sinker == nil {
		logger.LOGE("sinker no seted")
		return errors.New("no sinker to stop")
	}
	msg = &wssAPI.Msg{}
	msg.Type = wssAPI.MSG_STREAM_STOP
	err = this.sinker.ProcessMessage(msg)
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

func (this *streamSink) Id() string {
	return this.id
}
