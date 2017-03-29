package webSocketService

const (
	WS_status_ok = 200
)

//1byte type,
const (
	WS_pkt_audio   = 8
	WS_pkt_video   = 9
	WS_pkt_control = 18
)

//1byte control type
const (
	WS_ctrl_connect    = 1
	WS_ctrl_result     = 2
	WS_ctrl_play       = 3
	WS_ctrl_pause      = 4
	WS_ctrl_resume     = 5
	ws_ctrl_close      = 6
	WS_ctrl_publish    = 7
	WS_ctrl_onMetaData = 8
	WS_ctrl_unPublish  = 9
	WS_ctrl_stopPlay   = 10
	WS_ctrl_play2      = 11
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
