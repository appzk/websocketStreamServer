package webSocketService

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
	WS_ctrl_play2      = 3
	WS_ctrl_pause      = 4
	WS_ctrl_resume     = 5
	ws_ctrl_close      = 6
	ws_ctrl_send       = 7
	WS_ctrl_onMetaData = 8
	WS_ctrl_publish    = 9
	WS_ctrl_stop       = 10
)

type WsConnect struct {
	ID int `json:"id"`
}
