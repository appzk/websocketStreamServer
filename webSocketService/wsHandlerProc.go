package webSocketService

import (
	"encoding/json"
	"logger"
	"wssAPI"
)

import (
	"github.com/gorilla/websocket"
)

//add sink
func (this *websocketHandler) msg_play(stPlay *WsPlay) (err error) {
	logger.LOGT("play " + stPlay.StreamName)
	this.streamName = stPlay.StreamName
	this.playName = this.app + "/" + this.streamName
	this.clientId = wssAPI.GenerateGUID()
	err = this.addSink(this.playName, this.clientId)
	if err != nil {
		err = this.result(stPlay.ID, WS_status_notfound, this.playName+" not found")
		return nil
	}

	return this.result(stPlay.ID, WS_status_ok, "")
}

//stop play,del sink
func (this *websocketHandler) msg_stopPlay(stStopPlay *WsStopPlay) (err error) {
	logger.LOGT("stop play")
	this.stopPlay()
	this.delSink(this.playName, this.clientId)
	return this.result(stStopPlay.ID, WS_status_ok, "")
}

//del sink or source
func (this *websocketHandler) msg_close(stClose *WsClose) (err error) {
	this.stopPlay()
	this.stopPublish()
	this.delSink(this.playName, this.clientId)
	this.delSource(this.pubName)

	if stClose != nil {
		return this.result(stClose.ID, WS_status_ok, "")
	}
	return
}

func (this *websocketHandler) msg_connect(stConnect *WsConnect) (err error) {
	this.app = stConnect.App
	logger.LOGT("connect:" + this.app)
	return this.result(stConnect.ID, WS_status_ok, "")
}

func (this *websocketHandler) msg_result(stResult *WsResult) (err error) {
	return
}

//add source,start publish
func (this *websocketHandler) msg_publish(stPub *WsPublish) (err error) {
	return
}

//del source,stop publish
func (this *websocketHandler) msg_unPublish(stUnPub *WsUnPublish) (err error) {
	return
}

func (this *websocketHandler) sendStreamBegin() {
	stBegin := &WsStreamBegin{0}
	jsonData, _ := json.Marshal(stBegin)
	dataSend := make([]byte, len(jsonData)+2)
	dataSend[0] = WS_pkt_control
	dataSend[1] = WS_ctrl_streamBegin
	copy(dataSend[2:], jsonData)
	this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
}

func (this *websocketHandler) sendStreamEnd() {
	stBegin := &WsStreamEnd{0}
	jsonData, _ := json.Marshal(stBegin)
	dataSend := make([]byte, len(jsonData)+2)
	dataSend[0] = WS_pkt_control
	dataSend[1] = WS_ctrl_streamEnd
	copy(dataSend[2:], jsonData)
	this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
}
