package webSocketService

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"logger"
	"streamer"
	"sync"
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
	clientId     string
	isPlaying    bool
	mutexPlaying sync.RWMutex
	waitPlaying  *sync.WaitGroup
	stPlay       playInfo
	isPublish    bool
	mutexPublish sync.RWMutex
	hasSink      bool
	hasSource    bool
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
	this.stopPlay()
	if this.hasSink {
		streamer.DelSink(this.playName, this.clientId)
	}
	if this.hasSource{
		streamer.DelSource()
	}
	return
}

func (this *websocketHandler) GetType() string {
	return wsHandler
}

func (this *websocketHandler) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *websocketHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
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
				return
			}
			this.app = stConnect.App
			logger.LOGT("connect:" + this.app)
			return this.result(stConnect.ID, WS_status_ok, "")
		case WS_ctrl_result:
		case WS_ctrl_play:
			stPlay := &WsPlay{}
			err = json.Unmarshal(data[2:], stPlay)
			if err != nil {
				logger.LOGE("unmarshal json failed")
				return
			}
			this.streamName = stPlay.StreamName
			this.playName = this.app + "/" + this.streamName
			err = streamer.AddSink(this.playName, this.clientId, this)
			if err != nil {
				err = this.result(stPlay.ID, WS_status_notfound, this.playName+" not found")
				return nil
			}
			this.startPlay()
			return this.result(stPlay.ID, WS_status_ok, "")
		case WS_ctrl_play2:
		case WS_ctrl_pause:
		case WS_ctrl_resume:
		case ws_ctrl_close:
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
			this.stopPlay()
			return this.result(stStop.ID, WS_status_ok, "")
		}

	default:
		err = errors.New(fmt.Sprintf("msg type %d not supported", msgType))
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
	go this.threadPlay()
}

func (this *websocketHandler) stopPlay() {
	this.isPlaying = false
	this.waitPlaying.Wait()
}

func (this *websocketHandler) threadPlay() {
	this.waitPlaying.Add(1)
	defer func() {
		this.stPlay.reset()
		this.waitPlaying.Done()
	}()

	for this.isPlaying {

	}
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
