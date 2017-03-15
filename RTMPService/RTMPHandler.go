package RTMPService

import (
	"errors"
	"fmt"
	"logger"
	"wssAPI"
)

const (
	rtmp_status_idle       = 0
	rtmp_status_playing    = 1
	rtmp_status_publishing = 2
)

type RTMPHandler struct {
	status       int
	rtmpInstance *RTMP
	source       wssAPI.Obj
	sinke        wssAPI.Obj
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
	case RTMP_PACKET_TYPE_VIDEO:
	case RTMP_PACKET_TYPE_INFO:
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
		logger.LOGI(method.Value)
	default:
		logger.LOGW(fmt.Sprintf("rtmp method <%s> not processed", method.Value.StrValue))
	}
	return
}
