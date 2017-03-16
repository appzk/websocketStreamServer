package RTMPService

import (
	"errors"
	"fmt"
	"logger"
	"streamer"
	"sync"
	"wssAPI"
)

const (
	rtmp_status_idle       = 0
	rtmp_status_playing    = 1
	rtmp_status_publishing = 2
)

type RTMPHandler struct {
	mutexStatus  sync.RWMutex
	status       int
	rtmpInstance *RTMP
	source       wssAPI.Obj
	sinke        wssAPI.Obj
	streamName   string
	clientId     string
}

func (this *RTMPHandler) Init(msg *wssAPI.Msg) (err error) {
	this.status = rtmp_status_idle
	this.rtmpInstance = msg.Param1.(*RTMP)
	return
}

func (this *RTMPHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPHandler) Stop(msg *wssAPI.Msg) (err error) {
	this.mutexStatus.Lock()
	defer this.mutexStatus.Unlock()
	//stop publish or play
	//change status
	if rtmp_status_playing == this.status {
		this.stopPlaying()
		this.status = rtmp_status_idle
	} else if rtmp_status_publishing == this.status {
		this.stopPublishing()
		this.status = rtmp_status_idle
	}
	return
}

func (this *RTMPHandler) GetType() string {
	return rtmpTypeHandler
}

func (this *RTMPHandler) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *RTMPHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {

	if msg == nil {
		return errors.New("nil message")
	}

	return
}

func (this *RTMPHandler) Status() int {
	return this.status
}

func (this *RTMPHandler) HandleRTMPPacket(packet *RTMPPacket) (err error) {
	if nil == packet {
		this.shutdown()
		return
	}
	switch packet.MessageTypeId {
	case RTMP_PACKET_TYPE_CHUNK_SIZE:
		this.rtmpInstance.ChunkSize, err = AMF0DecodeInt32(packet.Body)
	case RTMP_PACKET_TYPE_CONTROL:
		err = this.rtmpInstance.HandleControl(packet)
	case RTMP_PACKET_TYPE_SERVER_BW:
		this.rtmpInstance.TargetBW, err = AMF0DecodeInt32(packet.Body)
	case RTMP_PACKET_TYPE_CLIENT_BW:
		this.rtmpInstance.SelfBW, err = AMF0DecodeInt32(packet.Body)
		this.rtmpInstance.LimitType = uint32(packet.Body[4])
	case RTMP_PACKET_TYPE_FLEX_MESSAGE:
		err = this.handleInvoke(packet)
	case RTMP_PACKET_TYPE_INVOKE:
		err = this.handleInvoke(packet)
	case RTMP_PACKET_TYPE_AUDIO:
		if this.status == rtmp_status_publishing && this.source != nil {

		} else {
			logger.LOGE("bad status")
			logger.LOGE(this.status)
			logger.LOGE(this.source)
		}
	case RTMP_PACKET_TYPE_VIDEO:
		if this.status == rtmp_status_publishing && this.source != nil {

		} else {
			logger.LOGE("bad status")
		}
	case RTMP_PACKET_TYPE_INFO:
		if this.status == rtmp_status_publishing && this.source != nil {

		} else {
			logger.LOGE("bad status")
		}
	default:
		logger.LOGW(fmt.Sprintf("rtmp packet type %d not processed", packet.MessageTypeId))
	}
	return
}

func (this *RTMPHandler) handleInvoke(packet *RTMPPacket) (err error) {
	var amfobj *AMF0Object
	if RTMP_PACKET_TYPE_FLEX_MESSAGE == packet.MessageTypeId {
		amfobj, err = AMF0DecodeObj(packet.Body[1:])
	} else {
		amfobj, err = AMF0DecodeObj(packet.Body)
	}
	if err != nil {
		logger.LOGE("recved invalid amf0 object")
		return
	}

	method := amfobj.Props.Front().Value.(*AMF0Property)

	switch method.Value.StrValue {
	case "connect":
		err = this.rtmpInstance.AcknowledgementBW()
		if err != nil {
			return
		}
		err = this.rtmpInstance.SetPeerBW()
		if err != nil {
			return
		}
		err = this.rtmpInstance.SetChunkSize(RTMP_better_chunk_size)
		if err != nil {
			return
		}
		err = this.rtmpInstance.OnBWDone()
		if err != nil {
			return
		}
		err = this.rtmpInstance.ConnectResult(amfobj)
		if err != nil {
			return
		}
	case "_checkbw":
		err = this.rtmpInstance.OnBWCheck()
	case "_result":
		this.handle_result(amfobj)
	case "releaseStream":
		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
			fmt.Sprintf("Method not found (%s).", "releaseStream"), idx)
	case "FCPublish":
		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
			fmt.Sprintf("Method not found (%s).", "FCPublish"), idx)
	case "createStream":
		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		err = this.rtmpInstance.CmdNumberResult(idx, 1.0)
	case "publish":

		//check prop
		if amfobj.Props.Len() < 4 {
			logger.LOGE("invalid props length")
			err = errors.New("invalid amf obj for publish")
			return
		}

		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		//check status
		if this.status != rtmp_status_idle {
			logger.LOGE("publish on bad status ")
			idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
			err = this.rtmpInstance.CmdError("error", "NetStream.Publish.Denied",
				fmt.Sprintf("can not publish (%s).", "publish"), idx)
			return
		}
		//add to source
		this.streamName = amfobj.AMF0GetPropByIndex(4).Value.StrValue + "/" +
			amfobj.AMF0GetPropByIndex(3).Value.StrValue
		this.source, err = streamer.Addsource(this.streamName)
		if err != nil {
			logger.LOGE("add source failed:" + err.Error())
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Publish.BadName",
				fmt.Sprintf("publish %s.", this.streamName), "", 0, RTMP_channel_Invoke)
			return
		}
		this.rtmpInstance.Link.Path = amfobj.AMF0GetPropByIndex(2).Value.StrValue
		err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
		if err != nil {
			logger.LOGE(err.Error())
			streamer.DelSource(this.streamName)
			return nil
		}
		err = this.startPublishing()
		if err != nil {
			logger.LOGE(err.Error())
			streamer.DelSource(this.streamName)
			return nil
		}
		this.status = rtmp_status_publishing
	case "FCUnpublish":
		this.shutdown()
	case "deleteStream":
		//do nothing now
	default:
		logger.LOGW(fmt.Sprintf("rtmp method <%s> not processed", method.Value.StrValue))
	}
	return
}

func (this *RTMPHandler) handle_result(amfobj *AMF0Object) {
	transactionId := int32(amfobj.AMF0GetPropByIndex(1).Value.NumValue)
	resultMethod := this.rtmpInstance.methodCache[transactionId]
	switch resultMethod {
	case "_onbwcheck":
	default:
		logger.LOGW("result of " + resultMethod + " not processed")
	}
}

func (this *RTMPHandler) shutdown() {
	this.mutexStatus.Lock()
	defer this.mutexStatus.Unlock()
	switch this.status {
	case rtmp_status_idle:
	case rtmp_status_publishing:
		this.stopPlaying()
		err := streamer.DelSource(this.streamName)
		if err != nil {
			logger.LOGE("del source failed" + err.Error())
		}
		this.status = rtmp_status_idle
	case rtmp_status_playing:
		this.stopPublishing()
		err := streamer.DelSink(this.streamName, this.clientId)
		if err != nil {
			logger.LOGE("del sink failed:" + err.Error())
		}
		this.status = rtmp_status_idle
	}
}

func (this *RTMPHandler) startPlaying() (err error) {
	return
}

func (this *RTMPHandler) stopPlaying() {

}

func (this *RTMPHandler) startPublishing() (err error) {
	err = this.rtmpInstance.CmdStatus("status", "NetStream.Publish.Start",
		fmt.Sprintf("publish %s", this.rtmpInstance.Link.Path), "", 0, RTMP_channel_Invoke)
	return
}

func (this *RTMPHandler) stopPublishing() (err error) {
	err = this.rtmpInstance.CmdStatus("status", "NetStream.Unpublish.Succes",
		fmt.Sprintf("unpublish %s", this.rtmpInstance.Link.Path), "", 0, RTMP_channel_Invoke)
	return
}
