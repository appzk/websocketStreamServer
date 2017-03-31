package webSocketService

import (
	"github.com/gorilla/websocket"
)

const (
	WS_status_ok       = 200
	WS_status_notfound = 404
)

//1byte type,
const (
	WS_pkt_audio   = 8
	WS_pkt_video   = 9
	WS_pkt_control = 18
)

//1byte control type
const (
	WS_ctrl_connect = iota
	WS_ctrl_result
	WS_ctrl_play
	WS_ctrl_pause
	WS_ctrl_resume
	ws_ctrl_close
	WS_ctrl_publish
	WS_ctrl_onMetaData
	WS_ctrl_unPublish
	WS_ctrl_stopPlay
	WS_ctrl_play2
	WS_ctrl_streamBegin
	WS_ctrl_streamEnd
)

type WsConnect struct {
	ID          int    `json:"id"`
	App         string `json:"app"`
	TcUrl       string `json:"tcUrl"`
	AudioCodecs int    `json:"audioCodecs"`
	VideoCodecs int    `json:"videoCodecs"`
}

type WsResult struct {
	ID     int    `json:"id"`
	Status int    `json:"status"`
	Desc   string `json:"desc"`
}

type WsPlay struct {
	ID         int     `json:"id"`
	StreamName string  `json:"streamName"`
	Start      float32 `json:"start"`
	Duration   float32 `json:"duration"`
	Reset      bool    `json:"reset"`
}

type WsPlay2 struct {
}

type WsPause struct {
	ID           int  `json:"id"`
	PauseFlag    bool `json:"pauseFlag"`
	MilliSeconds int  `json:"milliSeconds"`
}

type WsResume struct {
	ID int `json:"id"`
}

type WsClose struct {
	ID int `json:"id"`
}

type WsPublish struct {
	ID          int    `json:"id"`
	StreamName  string `json:"streamName"`
	PublishType string `json:"publishType"`
}

type WsOnMetadata struct {
}

type WsUnPublish struct {
	ID int `json:"id"`
}

type WsStopPlay struct {
	ID int `json:"id"`
}

type WsStreamBegin struct {
	ID int `json:"id"`
}

type WsStreamEnd struct {
	ID int `json:"id"`
}

func SendWsControl(conn *websocket.Conn, ctrlType int, data []byte) (err error) {
	dataSend := make([]byte, len(data)+2)
	dataSend[0] = WS_pkt_control
	dataSend[1] = byte(ctrlType)
	copy(dataSend[2:], data)
	return conn.WriteMessage(websocket.BinaryMessage, dataSend)
}
