package commonObj

const (
	OBJ_ServerBus       = "ServerBus"
	OBJ_RTMPServer      = "RTMPServer"
	OBJ_RTMPSource      = "RTMPSource"
	OBJ_RTMPSink        = "RTMPSink"
	OBJ_WebSocketServer = "WebsocketServer"
	OBJ_WebSocketSource = "WebsocketSource"
	OBJ_WebSocketSink   = "WebsocketSink"
)

const (
	TASK_CheckSource = "CheckSource" //param:source name
	TASK_AddSource   = "AddSource"   //param:source name  param2:sourceObj
	TASK_DelSource   = "DelSource"   //param:source name
	TASK_CheckSink   = "CheckSink"   //param:sink name
	TASK_AddSink     = "AddSink"     //param:sink name param2:sinkObj
	TASK_DelSink     = "DelSink"     //param:sink name
)

const (
	MSG_FLV_TAG = "FLVTag"
)
