package RTMPService

import (
	"container/list"
	"errors"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"streamer"
	"sync"
	"time"
	"wssAPI"
)

const (
	rtmp_status_idle          = 0
	rtmp_status_beforePlay    = 1 //recved play,not return play.start:start playing,stop nothing
	rtmp_status_playing       = 2 //return play.start,start send thread:start nothing,stop playing
	rtmp_status_beforePublish = 4 //recved publish:start publish,stop nil
	rtmp_status_publishing    = 5 //publishing:start nil,stop publish
	rtmp_status_closed        = 7
	//shutdown or play to publish or publish to play ,del
)

type RTMPHandler struct {
	mutexStatus  sync.RWMutex
	status       int
	rtmpInstance *RTMP
	source       wssAPI.Obj
	sinke        wssAPI.Obj
	srcAdded     bool
	sinkAdded    bool
	streamName   string
	clientId     string
	playInfo     RTMPPlayInfo
	app          string
}
type RTMPPlayInfo struct {
	playReset      bool
	playing        bool //true for thread send playing data
	waitPlaying    *sync.WaitGroup
	mutexCache     sync.RWMutex
	cache          *list.List
	audioHeader    *flv.FlvTag
	videoHeader    *flv.FlvTag
	metadata       *flv.FlvTag
	keyFrameWrited bool
	beginTime      uint32
	startTime      float32
	duration       float32
	reset          bool
}

func (this *RTMPHandler) Init(msg *wssAPI.Msg) (err error) {
	this.status = rtmp_status_idle
	this.rtmpInstance = msg.Param1.(*RTMP)
	this.playInfo.waitPlaying = new(sync.WaitGroup)
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
	switch msg.Type {
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		switch tag.TagType {
		case flv.FLV_TAG_Audio:
			if this.playInfo.audioHeader == nil {
				this.playInfo.audioHeader = tag
				this.playInfo.audioHeader.Timestamp = 0
				return
			}
		case flv.FLV_TAG_Video:
			if this.playInfo.videoHeader == nil {
				this.playInfo.videoHeader = tag
				this.playInfo.videoHeader.Timestamp = 0
				return
			}
			if false == this.playInfo.keyFrameWrited {
				if (tag.Data[0] >> 4) == 1 {
					this.playInfo.keyFrameWrited = true
					this.playInfo.beginTime = tag.Timestamp
					this.playInfo.sendInitPackets()
				} else {
					return
				}
			}

		case flv.FLV_TAG_ScriptData:
			if this.playInfo.metadata == nil {
				this.playInfo.metadata = tag
				//				obj, err := AMF0DecodeObj(tag.Data)

				//				if err == nil {
				//					obj.Dump()
				//					for v := obj.Props.Front(); v != nil; v = v.Next() {
				//						vp := v.Value.(*AMF0Property)
				//						if vp.PropType == AMF0_object {
				//							enc := &AMF0Encoder{}
				//							enc.Init()
				//							enc.EncodeAMFObj(&vp.Value.ObjValue)
				//							metadata, _ := enc.GetData()
				//							this.rtmpInstance.OnMetadata(metadata)
				//							obj, _ = AMF0DecodeObj(metadata)
				//							logger.LOGT(string(metadata))
				//							obj.Dump()
				//							break
				//						}
				//					}
				//				}

				return nil
			}
		}
		if false == this.playInfo.keyFrameWrited {
			return
		}
		tag.Timestamp -= this.playInfo.beginTime
		this.playInfo.mutexCache.Lock()
		this.playInfo.cache.PushBack(tag)
		this.playInfo.mutexCache.Unlock()
	case wssAPI.MSG_PLAY_START:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		logger.LOGT("start play")
		err = this.updateStatus(rtmp_status_playing)
		if err != nil {
			logger.LOGE("start play failed")
			return
		}

		err = this.startPlaying()
		return
	case wssAPI.MSG_PLAY_STOP:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		err = this.updateStatus(rtmp_status_beforePlay)
		if err != nil {
			logger.LOGE("stop playing failed")
			return
		}
		logger.LOGT("stop sink end")
		return
	case wssAPI.MSG_PUBLISH_START:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		err = this.updateStatus(rtmp_status_publishing)
		if err != nil {
			logger.LOGE("start publish failed")
			return
		}
		err = this.startPublishing()
		return
	case wssAPI.MSG_PUBLISH_STOP:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		err = this.updateStatus(rtmp_status_beforePublish)
		if err != nil {
			logger.LOGE("stop publish failed")
			return
		}
		err = this.stopPublishing()
		return
	default:
		logger.LOGW(fmt.Sprintf("msg type: %s not processed", msg.Type))
		return
	}
	return
}

func (this *RTMPHandler) Status() int {
	return this.status
}

func (this *RTMPHandler) HandleRTMPPacket(packet *RTMPPacket) (err error) {
	if nil == packet {
		this.updateStatus(rtmp_status_idle)
		return
	}
	switch packet.MessageTypeId {
	case RTMP_PACKET_TYPE_CHUNK_SIZE:
		this.rtmpInstance.RecvChunkSize, err = AMF0DecodeInt32(packet.Body)
		logger.LOGT(fmt.Sprintf("chunk size:%d", this.rtmpInstance.RecvChunkSize))
	case RTMP_PACKET_TYPE_CONTROL:
		err = this.rtmpInstance.HandleControl(packet)
	case RTMP_PACKET_TYPE_BYTES_READ_REPORT:
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
			msg := &wssAPI.Msg{}
			msg.Type = wssAPI.MSG_FLV_TAG
			msg.Param1 = packet.ToFLVTag()
			this.source.ProcessMessage(msg)
		} else {
			logger.LOGE("bad status")
			logger.LOGE(this.status)
			logger.LOGE(this.source)
		}
	case RTMP_PACKET_TYPE_VIDEO:
		if this.status == rtmp_status_publishing && this.source != nil {
			msg := &wssAPI.Msg{}
			msg.Type = wssAPI.MSG_FLV_TAG
			msg.Param1 = packet.ToFLVTag()
			this.source.ProcessMessage(msg)
		} else {
			logger.LOGE("bad status")
		}
	case RTMP_PACKET_TYPE_INFO:
		if this.status == rtmp_status_publishing && this.source != nil {
			msg := &wssAPI.Msg{}
			msg.Type = wssAPI.MSG_FLV_TAG
			//logger.LOGI(packet.ChunkStreamID)
			msg.Param1 = packet.ToFLVTag()
			this.source.ProcessMessage(msg)
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
	if amfobj.Props.Len() == 0 {
		logger.LOGT(packet.Body)
		logger.LOGT(string(packet.Body))
		return
	}

	method := amfobj.Props.Front().Value.(*AMF0Property)

	switch method.Value.StrValue {
	case "connect":
		cmdObj := amfobj.AMF0GetPropByIndex(2)
		if cmdObj != nil {
			this.app = cmdObj.Value.ObjValue.AMF0GetPropByName("app").Value.StrValue
		}
		if this.app != serviceConfig.LivePath {
			logger.LOGW("path wrong")
		}
		err = this.rtmpInstance.AcknowledgementBW()
		if err != nil {
			return
		}
		err = this.rtmpInstance.SetPeerBW()
		if err != nil {
			return
		}
		//err = this.rtmpInstance.SetChunkSize(RTMP_better_chunk_size)
		//		if err != nil {
		//			return
		//		}
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
		//		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		//		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
		//			fmt.Sprintf("Method not found (%s).", "releaseStream"), idx)
	case "FCPublish":
		//		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		//		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
		//			fmt.Sprintf("Method not found (%s).", "FCPublish"), idx)
	case "createStream":
		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		err = this.rtmpInstance.CmdNumberResult(idx, 1.0)
	case "publish":
		this.updateStatus(rtmp_status_beforePublish)
		//check prop
		if amfobj.Props.Len() < 4 {
			logger.LOGE("invalid props length")
			err = errors.New("invalid amf obj for publish")
			return
		}

		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		//check status
		if this.status != rtmp_status_beforePublish {
			logger.LOGE("publish on bad status ")
			idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
			err = this.rtmpInstance.CmdError("error", "NetStream.Publish.Denied",
				fmt.Sprintf("can not publish (%s).", "publish"), idx)
			return
		}
		//add to source
		this.streamName = this.app + "/" + amfobj.AMF0GetPropByIndex(3).Value.StrValue
		this.source, err = streamer.Addsource(this.streamName)
		if err != nil {
			logger.LOGE("add source failed:" + err.Error())
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Publish.BadName",
				fmt.Sprintf("publish %s.", this.streamName), "", 0, RTMP_channel_Invoke)
			this.streamName = ""
			this.updateStatus(rtmp_status_idle)
			return errors.New("bad name")
		}
		this.srcAdded = true
		this.rtmpInstance.Link.Path = amfobj.AMF0GetPropByIndex(2).Value.StrValue
		err = this.startPublishing()
		if err != nil {
			logger.LOGE(err.Error())
			this.updateStatus(rtmp_status_idle)
			return nil
		}
		this.updateStatus(rtmp_status_publishing)
	case "FCUnpublish":
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		this.updateStatus(rtmp_status_idle)
	case "deleteStream":
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		this.updateStatus(rtmp_status_idle)
	//do nothing now
	case "play":
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		this.updateStatus(rtmp_status_beforePlay)
		this.streamName = this.app + "/" + amfobj.AMF0GetPropByIndex(3).Value.StrValue
		this.rtmpInstance.Link.Path = this.streamName
		this.playInfo.startTime = -2
		this.playInfo.duration = -1
		this.playInfo.reset = false
		if amfobj.Props.Len() >= 5 {
			this.playInfo.startTime = float32(amfobj.AMF0GetPropByIndex(4).Value.NumValue)
		}
		if amfobj.Props.Len() >= 6 {
			this.playInfo.duration = float32(amfobj.AMF0GetPropByIndex(5).Value.NumValue)
			if this.playInfo.duration < 0 {
				this.playInfo.duration = -1
			}
		}
		if amfobj.Props.Len() >= 7 {
			this.playInfo.reset = amfobj.AMF0GetPropByIndex(6).Value.BoolValue
		}
		err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}

		if true == this.playInfo.playReset {
			err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Reset",
				fmt.Sprintf("Playing and resetting %s", this.rtmpInstance.Link.Path),
				this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		}

		err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Start",
			fmt.Sprintf("Started playing %s", this.rtmpInstance.Link.Path), this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		this.clientId = wssAPI.GenerateGUID()
		this.mutexStatus.Unlock()
		err = streamer.AddSink(this.streamName, this.clientId, this)
		if err != nil {
			//404
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Play.StreamNotFound",
				"paly failed", this.streamName, 0, RTMP_channel_Invoke)
			this.mutexStatus.Lock()
			return nil
		}
		this.sinkAdded = true
		this.mutexStatus.Lock()
	case "_error":
		amfobj.Dump()
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

//onstatus publish notify
func (this *RTMPHandler) startPlaying() (err error) {
	this.playInfo.waitPlaying.Wait()
	err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.PublishNotify",
		fmt.Sprintf("%s is now unpublished", this.rtmpInstance.Link.Path),
		this.rtmpInstance.Link.Path,
		0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
	}
	err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	if true == this.playInfo.playReset {
		err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Reset",
			fmt.Sprintf("Playing and resetting %s", this.rtmpInstance.Link.Path),
			this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
	}

	err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Start",
		fmt.Sprintf("Started playing %s", this.rtmpInstance.Link.Path), this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
		return
	}

	logger.LOGT("start playing")

	go this.threadPlaying()
	return
}

//fcunpublish
//onstatus  unpublish notify
//no play.stop
func (this *RTMPHandler) stopPlaying() (err error) {
	//stop playing thread
	this.playInfo.playing = false
	this.playInfo.audioHeader = nil
	this.playInfo.videoHeader = nil
	this.playInfo.metadata = nil
	this.playInfo.beginTime = 0
	this.playInfo.keyFrameWrited = false
	this.playInfo.waitPlaying.Wait()

	err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamEof, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		//logger.LOGT("stop play")
		//streamer.DelSink(this.streamName, this.clientId)
		return nil
	}

	err = this.rtmpInstance.FCUnpublish()
	if err != nil {
		logger.LOGE("FCUnpublish failed:" + err.Error())
		return
	}

	err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.UnpublishNotify",
		fmt.Sprintf("%s is unpublished", this.rtmpInstance.Link.Path),
		this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)

	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	//err=this.rtmpInstance.
	//	err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Stop",
	//		fmt.Sprintf("Stoped playing %s", this.rtmpInstance.Link.Path), this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
	//	if err != nil {
	//		logger.LOGE(err.Error())
	//		return
	//	}
	logger.LOGT("stop play")
	return
}

func (this *RTMPHandler) startPublishing() (err error) {
	err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return nil
	}
	err = this.rtmpInstance.CmdStatus("status", "NetStream.Publish.Start",
		fmt.Sprintf("publish %s", this.rtmpInstance.Link.Path), "", 0, RTMP_channel_Invoke)
	return
}

func (this *RTMPHandler) stopPublishing() (err error) {

	err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamEof, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return nil
	}
	err = this.rtmpInstance.CmdStatus("status", "NetStream.Unpublish.Succes",
		fmt.Sprintf("unpublish %s", this.rtmpInstance.Link.Path), "", 0, RTMP_channel_Invoke)
	return
}

//改变状态，移除源或槽
func (this *RTMPHandler) updateStatus(status int) (err error) {

	switch status {
	case rtmp_status_idle:
		return this.updateToIdle()
	case rtmp_status_beforePublish:
		return this.updateToBeforePublish()
	case rtmp_status_publishing:
		return this.updateToPublishing()
	case rtmp_status_beforePlay:
		return this.updateToBeforePlay()
	case rtmp_status_playing:
		return this.updateToPlaying()
	}

	return errors.New(fmt.Sprintf("update status error"))
}

func (this *RTMPHandler) threadPlaying() {
	this.playInfo.playing = true
	this.playInfo.waitPlaying.Add(1)
	defer func() {
		logger.LOGT("thread play end")
		this.playInfo.cache = list.New()
		this.playInfo.waitPlaying.Done()
	}()

	for this.playInfo.playing == true {
		this.playInfo.mutexCache.Lock()
		if this.playInfo.cache == nil || this.playInfo.cache.Len() == 0 {
			this.playInfo.mutexCache.Unlock()
			time.Sleep(30 * time.Millisecond)
			continue
		}
		if this.playInfo.cache.Len() > serviceConfig.CacheCount {
			this.playInfo.mutexCache.Unlock()
			//bw not enough
			this.rtmpInstance.CmdStatus("warning", "NetStream.Play.InsufficientBW",
				"instufficient bw", this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
			//shutdown
			this.updateStatus(rtmp_status_idle)
			return
		}
		tag := this.playInfo.cache.Front().Value.(*flv.FlvTag)
		this.playInfo.cache.Remove(this.playInfo.cache.Front())
		this.playInfo.mutexCache.Unlock()
		//时间错误

		err := this.rtmpInstance.SendPacket(FlvTagToRTMPPacket(tag), false)
		if err != nil {
			this.updateStatus(rtmp_status_idle)
			logger.LOGE("send rtmp packet failed in play")
			return
		}
	}
}

func (this *RTMPPlayInfo) sendInitPackets() {
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()

	//logger.LOGT("send init packet")
	if this.cache == nil {
		this.cache = list.New()
	}

	if this.audioHeader != nil {
		this.cache.PushBack(this.audioHeader)
	}

	if this.videoHeader != nil {
		this.cache.PushBack(this.videoHeader)
	}

	if this.metadata != nil {
		this.cache.PushBack(this.metadata)
	}
}

func (this *RTMPHandler) updateToIdle() (err error) {

	switch this.status {
	case rtmp_status_beforePlay:
		err = streamer.DelSink(this.streamName, this.clientId)
		this.status = rtmp_status_idle
		return
	case rtmp_status_playing:
		err = this.stopPlaying()
		if err != nil {
			logger.LOGE("stop playing failed:" + err.Error())
		}
		err = streamer.DelSink(this.streamName, this.clientId)
		this.status = rtmp_status_idle
		return
	case rtmp_status_beforePublish:
		err = streamer.DelSource(this.streamName)
		this.status = rtmp_status_idle
		return
	case rtmp_status_publishing:
		err = this.stopPublishing()
		if err != nil {
			logger.LOGE("stop publish failed:" + err.Error())
		}
		err = streamer.DelSource(this.streamName)
		this.status = rtmp_status_idle
		return
	}
	return
}

func (this *RTMPHandler) updateToBeforePlay() (err error) {
	switch this.status {
	case rtmp_status_beforePlay:
		logger.LOGW("double before play:del sink")
		err = streamer.DelSink(this.streamName, this.clientId)
		this.status = rtmp_status_beforePlay
		return
	case rtmp_status_idle:
		this.status = rtmp_status_beforePlay
		return
	case rtmp_status_playing:
		err = this.stopPlaying()
		if err != nil {
			logger.LOGE("stop playing failed:" + err.Error())
		}
		//err = streamer.DelSink(this.streamName, this.clientId)
		this.status = rtmp_status_beforePlay
		return
	default:
		logger.LOGW(fmt.Sprintf("update status to beforePlay from %d not processed", this.status))
	}
	return
}

func (this *RTMPHandler) updateToPlaying() (err error) {

	if this.status == rtmp_status_beforePlay {
		this.status = rtmp_status_playing
		return
	}
	return
}

func (this *RTMPHandler) updateToBeforePublish() (err error) {

	switch this.status {
	case rtmp_status_beforePublish:
		this.status = rtmp_status_beforePublish
		return
	case rtmp_status_idle:
		this.status = rtmp_status_beforePublish
		return
	case rtmp_status_publishing:
		//stop publish
		err = this.stopPublishing()
		if err != nil {
			logger.LOGE("stop publish failed:" + err.Error())
		}
		err = streamer.DelSource(this.streamName)
		this.status = rtmp_status_beforePublish
		return
	default:
		logger.LOGW(fmt.Sprintf("update status to beforePublish from %d not processed", this.status))
	}
	return
}

func (this *RTMPHandler) updateToPublishing() (err error) {
	if this.status == rtmp_status_beforePublish {
		this.status = rtmp_status_publishing
		return
	}
	return
}

func (this *RTMPHandler) updateToClosed() (err error) {
	return
}
