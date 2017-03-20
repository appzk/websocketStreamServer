package wssAPI

const (
	OBJ_ServerBus       = "ServerBus"
	OBJ_RTMPServer      = "RTMPServer"
	OBJ_WebSocketServer = "WebsocketServer"
)

const (
	TASK_AddSource = "AddSource" //param:source name  param2:sourceObj
	TASK_DelSource = "DelSource" //param:source name
	TASK_AddSink   = "AddSink"   //param:sink name param2:sinkObj
	TASK_DelSink   = "DelSink"   //param:sink name
)

const (
	MSG_FLV_TAG       = "FLVTag"
	MSG_PUBLISH_START = "NetStream.Publish.Start"
	MSG_PUBLISH_STOP  = "NetStream.Publish.Stop"
	MSG_PLAY_START    = "NetStream.Play.Start"
	MSG_PLAY_STOP     = "NetStream.Play.Stop"
)
