package webSocketService

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"mediaTypes/mp4"
	"streamer"
	"sync"
	"time"
	"wssAPI"

	"github.com/gorilla/websocket"
)

const (
	wsHandler = "websocketHandler"
)

type websocketHandler struct {
	conn         *websocket.Conn
	app          string
	streamName   string
	playName     string
	pubName      string
	clientId     string
	isPlaying    bool
	mutexPlaying sync.RWMutex
	waitPlaying  *sync.WaitGroup
	stPlay       playInfo
	isPublish    bool
	mutexPublish sync.RWMutex
	hasSink      bool
	mutexbSink   sync.RWMutex
	hasSource    bool
	mutexbSource sync.RWMutex
	source       wssAPI.Obj
}

type playInfo struct {
	cache          *list.List
	mutexCache     sync.RWMutex
	audioHeader    *flv.FlvTag
	videoHeader    *flv.FlvTag
	metadata       *flv.FlvTag
	keyFrameWrited bool
	beginTime      uint32
}

func (this *websocketHandler) Init(msg *wssAPI.Msg) (err error) {
	this.conn = msg.Param1.(*websocket.Conn)
	this.waitPlaying = new(sync.WaitGroup)
	return
}

func (this *websocketHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) Stop(msg *wssAPI.Msg) (err error) {
	this.msg_close(nil)
	return
}

func (this *websocketHandler) GetType() string {
	return wsHandler
}

func (this *websocketHandler) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *websocketHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		switch tag.TagType {
		case flv.FLV_TAG_Audio:
			if this.stPlay.audioHeader == nil {
				this.stPlay.audioHeader = tag
				this.stPlay.audioHeader.Timestamp = 0
				return
			}
		case flv.FLV_TAG_Video:
			if this.stPlay.videoHeader == nil {
				this.stPlay.videoHeader = tag
				this.stPlay.videoHeader.Timestamp = 0
				return
			}
			if false == this.stPlay.keyFrameWrited {
				if (tag.Data[0] >> 4) == 1 {
					this.stPlay.keyFrameWrited = true
					this.stPlay.beginTime = tag.Timestamp
					this.stPlay.addInitPkts()
				} else {
					return
				}
			}

		case flv.FLV_TAG_ScriptData:
			if this.stPlay.metadata == nil {
				this.stPlay.metadata = tag

				return nil
			}
		}
		if false == this.stPlay.keyFrameWrited {
			return
		}
		tag.Timestamp -= this.stPlay.beginTime
		this.stPlay.mutexCache.Lock()
		this.stPlay.cache.PushBack(tag)
		this.stPlay.mutexCache.Unlock()
	case wssAPI.MSG_PLAY_START:
		this.startPlay()
	case wssAPI.MSG_PLAY_STOP:
		this.stopPlay()
	case wssAPI.MSG_PUBLISH_START:
		this.startPublish()
	case wssAPI.MSG_PUBLISH_STOP:
		this.stopPublish()
	}
	return
}

func (this *websocketHandler) processWSMessage(data []byte) (err error) {
	if nil == data {
		this.Stop(nil)
		return
	}
	msgType := int(data[0])
	switch msgType {
	case WS_pkt_audio:
	case WS_pkt_video:
	case WS_pkt_control:
		ctrlType := int(data[1])
		switch ctrlType {
		case WS_ctrl_connect:
			stConnect := &WsConnect{}
			err = json.Unmarshal(data[2:], stConnect)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
			return this.msg_connect(stConnect)

		case WS_ctrl_result:
		case WS_ctrl_play:
			stPlay := &WsPlay{}
			err = json.Unmarshal(data[2:], stPlay)
			if err != nil {
				logger.LOGE("unmarshal json failed")
				return
			}
			return this.msg_play(stPlay)

		case WS_ctrl_play2:
		case WS_ctrl_pause:
		case WS_ctrl_resume:
		case ws_ctrl_close:
			stClose := &WsClose{}
			err = json.Unmarshal(data[2:], stClose)
			if err != nil {
				logger.LOGE("unmarshal json failed")
				return
			}
			return this.msg_close(stClose)
		case WS_ctrl_publish:
		case WS_ctrl_onMetaData:
		case WS_ctrl_unPublish:
		case WS_ctrl_stopPlay:
			stStop := &WsStopPlay{}
			err = json.Unmarshal(data[2:], stStop)
			if err != nil {
				logger.LOGE("unmarshal json stopplay failed")
				return
			}
			return this.msg_stopPlay(stStop)
		}

	default:
		err = errors.New(fmt.Sprintf("msg type %d not supported", msgType))
		logger.LOGW("invalid binary data")
		return
	}
	return
}

func (this *websocketHandler) result(id, status int, desc string) (err error) {
	stResult := &WsResult{id, status, desc}
	jsondata, err := json.Marshal(stResult)
	if err != nil {
		logger.LOGE("marshal json failed")
		return
	}
	dataSend := make([]byte, len(jsondata)+2)
	dataSend[0] = WS_pkt_control
	dataSend[1] = WS_ctrl_result
	copy(dataSend[2:], jsondata)
	err = this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
	return
}

func (this *websocketHandler) startPlay() {
	logger.LOGT("start play")
	//start play thread
	this.mutexPlaying.RLock()
	if this.isPlaying == true {
		this.mutexPlaying.RUnlock()
		return
	}
	this.mutexPlaying.RUnlock()
	this.waitPlaying.Wait()
	this.stPlay.reset()
	this.isPlaying = true
	//send streamBegin
	go this.threadPlay()
}

func (this *websocketHandler) stopPlay() {
	this.isPlaying = false
	this.waitPlaying.Wait()
	//send streamEnd
}

func (this *websocketHandler) threadPlay() {
	this.waitPlaying.Add(1)
	defer func() {
		this.stPlay.reset()
		this.waitPlaying.Done()
		this.sendStreamEnd()
	}()

	this.sendStreamBegin()
	fmp4Creater := &mp4.FMP4Creater{}
	for this.isPlaying {
		this.stPlay.mutexCache.Lock()
		if this.stPlay.cache == nil || this.stPlay.cache.Len() == 0 {
			this.stPlay.mutexCache.Unlock()
			time.Sleep(30 * time.Millisecond)
			continue
		}
		//flv tag to fmp4 pkt
		for v := this.stPlay.cache.Front(); v != nil; v = v.Next() {
			//tag := v.Value.(*flv.FlvTag)
			this.stPlay.cache.Remove(this.stPlay.cache.Front())
			slice := fmp4Creater.AddFlvTag(v.Value.(*flv.FlvTag))
			if slice != nil {
				this.sendSlice(slice)
			}
		}
		this.stPlay.mutexCache.Unlock()
	}
}

func (this *websocketHandler) startPublish() {

}

func (this *websocketHandler) stopPublish() {

}

func (this *websocketHandler) delSink(name, id string) (err error) {
	this.mutexbSink.Lock()
	defer this.mutexbSink.Unlock()
	if this.hasSink {
		err = streamer.DelSink(name, id)
		if err != nil {
			logger.LOGE("delete sink:" + name + " failed --" + id)
			logger.LOGE(err.Error())
		} else {

			logger.LOGT("del sink:" + id)
		}
		this.hasSink = false
	} else {
		err = errors.New("del not existed sink")
	}
	return
}

func (this *websocketHandler) addSink(name, id string) (err error) {
	this.mutexbSink.Lock()
	defer this.mutexbSink.Unlock()
	if false == this.hasSink {
		err = streamer.AddSink(name, id, this)
		if err != nil {
			logger.LOGE(fmt.Sprintf("add sink %s to %s failed", name, id))
			return
		}
		this.hasSink = true
	} else {
		logger.LOGE("sink not deleted")
		err = errors.New("sink not deleted")
	}
	return
}

func (this *websocketHandler) addSource(name string) (err error) {
	this.mutexbSource.Lock()
	defer this.mutexbSource.Unlock()
	if this.hasSource {
		logger.LOGE("source existed")
		return errors.New("add existed source")
	}
	this.source, err = streamer.Addsource(name)
	if err != nil {
		logger.LOGE("add source:" + name + " failed")
		return
	}
	this.hasSource = true
	return
}

func (this *websocketHandler) delSource(name string) (err error) {
	this.mutexbSource.Lock()
	defer this.mutexbSource.Unlock()
	if this.hasSource == false {
		return errors.New("del not existed source")
	}
	err = streamer.DelSource(name)
	if err != nil {
		logger.LOGE("del source:" + name + " failed")
		return
	}
	this.hasSource = false
	return
}

func (this *websocketHandler) sendSlice(slice *mp4.FMP4Slice) (err error) {
	dataSend := make([]byte, len(slice.Data)+1)
	dataSend[0] = byte(slice.Type)
	copy(dataSend[1:], slice.Data)
	return this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
}

func (this *playInfo) reset() {
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
	this.cache = list.New()
	this.audioHeader = nil
	this.videoHeader = nil
	this.metadata = nil
	this.keyFrameWrited = false
	this.beginTime = 0
}

func (this *playInfo) addInitPkts() {
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
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
