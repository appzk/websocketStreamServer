package RTMPService

import (
	"errors"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"net"
	"wssAPI"
)

const (
	RTMP_protocol_rtmp      = "rtmp"
	RTMP_default_chunk_size = 128
	RTMP_default_buff_ms    = 500
	RTMP_better_chunk_size  = 4096

	RTMP_channel_control      = 0x02
	RTMP_channel_Invoke       = 0x03
	RTMP_channel_SendLive     = 0x04
	RTMP_channel_SendPlayback = 0x05
	RTMP_channel_AV           = 0x08
)

const (
	RTMP_PACKET_TYPE_CHUNK_SIZE         = 0x01
	RTMP_PACKET_TYPE_ABORT              = 0x02
	RTMP_PACKET_TYPE_BYTES_READ_REPORT  = 0x03
	RTMP_PACKET_TYPE_CONTROL            = 0x04
	RTMP_PACKET_TYPE_SERVER_BW          = 0x05
	RTMP_PACKET_TYPE_CLIENT_BW          = 0x06
	RTMP_PACKET_TYPE_AUDIO              = 0x08
	RTMP_PACKET_TYPE_VIDEO              = 0x09
	RTMP_PACKET_TYPE_FLEX_STREAM_SEND   = 0x0f //amf3
	RTMP_PACKET_TYPE_FLEX_SHARED_OBJECT = 0x10
	RTMP_PACKET_TYPE_FLEX_MESSAGE       = 0x11
	RTMP_PACKET_TYPE_INFO               = 0x12 //amf0
	RTMP_PACKET_TYPE_SHARED_OBJECT      = 0x13
	RTMP_PACKET_TYPE_INVOKE             = 0x14
	RTMP_PACKET_TYPE_FLASH_VIDEO        = 0x16
)

const (
	RTMP_CTRL_streamBegin       = 0
	RTMP_CTRL_streamEof         = 1
	RTMP_CTRL_streamDry         = 2
	RTMP_CTRL_setBufferLength   = 3
	RTMP_CTRL_streamIsRecorded  = 4
	RTMP_CTRL_pingRequest       = 6
	RTMP_CTRL_pingResponse      = 7
	RTMP_CTRL_streamBufferEmpty = 31
	RTMP_CTRL_streamBufferReady = 32
)

type RTMP_LINK struct {
	Protocol    string
	App         string
	Flashver    string
	SwfUrl      string
	TcUrl       string
	Fpad        string
	AudioCodecs string
	VideoCodecs string
	Path        string
	Address     string
}

type RTMPPacket struct {
	ChunkStreamID   int32
	Fmt             byte
	TimeStamp       uint32 //fmt 0为绝对时间，1或2为相对时间
	MessageLength   uint32
	MessageTypeId   byte
	MessageStreamId uint32
	Body            []byte
	BodyReaded      int32
}

type RTMP struct {
	Conn        net.Conn
	ChunkSize   uint32
	NumInvokes  int32
	StreamId    uint32
	Link        RTMP_LINK
	AudioCodecs int32
	VideoCodecs int32
	TargetBW    uint32
	SelfBW      uint32
	LimitType   uint32
	BytesIn     int64
	BytesInLast int64
	buffMS      uint32
	recvCache   map[int32]*RTMPPacket
	methodCache map[int32]string
}

func (this *RTMPPacket) Copy() (dst *RTMPPacket) {
	dst = &RTMPPacket{}
	dst.BodyReaded = this.BodyReaded
	dst.Body = make([]byte, dst.BodyReaded)
	copy(dst.Body, this.Body)
	dst.ChunkStreamID = this.ChunkStreamID
	dst.Fmt = this.Fmt
	dst.MessageLength = this.MessageLength
	dst.MessageStreamId = this.MessageStreamId
	dst.TimeStamp = this.TimeStamp
	dst.MessageTypeId = this.MessageTypeId
	return
}

func (this *RTMPPacket) ToFLVTag() (dst *flv.FlvTag) {
	return
}

func FlvTagToRTMPPacket(ta *flv.FlvTag) (dst *RTMPPacket) {
	dst = &RTMPPacket{}
	dst.ChunkStreamID = RTMP_channel_SendLive
	dst.Fmt = 0
	dst.TimeStamp = ta.Timestamp
	dst.MessageStreamId = 1
	dst.MessageTypeId = ta.TagType
	dst.MessageLength = uint32(len(ta.Data))
	dst.Body = make([]byte, dst.MessageLength)
	copy(dst.Body, ta.Data)
	return
}

func (this *RTMP) Init(conn net.Conn) {
	this.Conn = conn
	this.ChunkSize = RTMP_default_chunk_size
	this.AudioCodecs = 3191
	this.VideoCodecs = 252
	this.TargetBW = 2500000
	this.SelfBW = 2500000
	this.LimitType = 2
	this.buffMS = RTMP_default_buff_ms
	this.recvCache = make(map[int32]*RTMPPacket)
	this.methodCache = make(map[int32]string)
}

func (this *RTMP) ReadPacket() (packet *RTMPPacket, err error) {
	for {
		packet, err = this.ReadChunk()
		if err != nil {
			logger.LOGT("read rtmp chunk failed")
			return
		}
		if nil != packet && packet.BodyReaded > 0 &&
			packet.BodyReaded == int32(len(packet.Body)) {
			return
		}
	}
	return
}

func timeAdd(src, delta uint32) (ret uint32) {
	ret = src + delta
	return
}

func (this *RTMP) ReadChunk() (packet *RTMPPacket, err error) {
	//接收basic header
	buf, err := wssAPI.TcpRead(this.Conn, 1)
	if err != nil {
		return
	}
	var fmt byte
	var chunkId int32
	fmt = buf[0] >> 6
	chunkId = int32(buf[0] & 0x3f)
	if chunkId == 0 {
		buf, err = wssAPI.TcpRead(this.Conn, 1)
		if err != nil {
			return
		}
		chunkId = int32(buf[0]) + 64
	} else if chunkId == 1 {
		buf, err = wssAPI.TcpRead(this.Conn, 2)
		if err != nil {
			return
		}
		tmp16 := int32(buf[0]) + int32(buf[1])*256 + 64
		if err != nil {
			return nil, err
		}
		chunkId = int32(tmp16)
	}
	if this.recvCache[chunkId] == nil {
		this.recvCache[chunkId] = &RTMPPacket{}
		this.recvCache[chunkId].Fmt = 0
	}
	this.recvCache[chunkId].ChunkStreamID = chunkId
	//接收message header
	switch fmt {
	case 0:
		buf, err := wssAPI.TcpRead(this.Conn, 11)
		if err != nil {
			return nil, err
		}
		tmpPkt := this.recvCache[chunkId]
		tmpPkt.TimeStamp, _ = AMF0DecodeInt24(buf)
		tmpPkt.MessageLength, _ = AMF0DecodeInt24(buf[3:])
		tmpPkt.MessageTypeId = buf[6]
		tmpPkt.MessageStreamId, _ = AMF0DecodeInt32LE(buf[7:])
		if tmpPkt.TimeStamp == 0xffffff {
			buf, err = wssAPI.TcpRead(this.Conn, 4)
			if err != nil {
				return nil, err
			}
			tmpPkt.TimeStamp, _ = AMF0DecodeInt32(buf)
		}
	case 1:
		buf, err := wssAPI.TcpRead(this.Conn, 7)
		if err != nil {
			return nil, err
		}
		tmpPkt := this.recvCache[chunkId]
		timeDelta, _ := AMF0DecodeInt24(buf)
		tmpPkt.MessageLength, _ = AMF0DecodeInt24(buf[3:])
		tmpPkt.MessageTypeId = buf[6]
		if timeDelta == 0xffffff {
			buf, err = wssAPI.TcpRead(this.Conn, 4)
			if err != nil {
				return nil, err
			}
			timeDelta, _ = AMF0DecodeInt32(buf)
		}
		tmpPkt.TimeStamp = timeAdd(tmpPkt.TimeStamp, timeDelta)
	case 2:
		buf, err := wssAPI.TcpRead(this.Conn, 3)

		if err != nil {
			return nil, err
		}

		tmpPkt := this.recvCache[chunkId]
		timeDelta, _ := AMF0DecodeInt24(buf)
		if timeDelta == 0xffffff {
			buf, err = wssAPI.TcpRead(this.Conn, 4)
			if err != nil {
				return nil, err
			}
			timeDelta, _ = AMF0DecodeInt32(buf)
		}
		tmpPkt.TimeStamp = timeAdd(tmpPkt.TimeStamp, timeDelta)

	case 3:

	}
	//接收chunk data
	tmpPkt := this.recvCache[chunkId]
	if tmpPkt.MessageLength == 0 {
		return
	}
	if fmt != 3 {
		tmpPkt.Body = make([]byte, tmpPkt.MessageLength)
		tmpPkt.BodyReaded = 0
	}
	//接收小于等于一个chunksize的数据
	recvsize := tmpPkt.MessageLength - uint32(tmpPkt.BodyReaded)
	if recvsize > this.ChunkSize {
		recvsize = this.ChunkSize
	}
	tmpBody, err := wssAPI.TcpRead(this.Conn, int(recvsize))
	if err != nil {
		return
	}
	copy(tmpPkt.Body[tmpPkt.BodyReaded:], tmpBody)
	tmpPkt.BodyReaded += int32(recvsize)
	//判断是否收到一个完整的包
	if tmpPkt.BodyReaded == int32(tmpPkt.MessageLength) {
		packet = tmpPkt.Copy()
	}
	return
}

func (this *RTMP) SendPacket(packet *RTMPPacket, queue bool) (err error) {
	//基本头
	encoder := &AMF0Encoder{}
	encoder.Init()
	var tmp8 byte
	if packet.ChunkStreamID < 63 {
		tmp8 = (packet.Fmt << 6) | byte(packet.ChunkStreamID)
		encoder.AppendByte(tmp8)
	} else if packet.ChunkStreamID < 319 {
		tmp8 = (packet.Fmt << 6)
		encoder.AppendByte(tmp8)
		encoder.AppendByte(byte(packet.ChunkStreamID - 64))
	} else if packet.ChunkStreamID < 65599 {
		tmp8 = (packet.Fmt << 6) | byte(1)
		encoder.AppendByte(tmp8)
		encoder.AppendByte(byte(packet.ChunkStreamID-64) & 0xff)
		encoder.AppendByte(byte(packet.ChunkStreamID-64) >> 8)
	} else {
		return errors.New(fmt.Sprintf("chunk stream id:%d invalid", packet.ChunkStreamID))
	}
	//消息头
	if packet.Fmt <= 2 {
		//只有时间
		if packet.TimeStamp >= 0xffffff {
			encoder.EncodeInt24(0xffffff)
		} else {
			encoder.EncodeInt24(int32(packet.TimeStamp))
		}
	}
	if packet.Fmt <= 1 {
		//有长度，类型
		encoder.EncodeInt24(int32(packet.MessageLength))
		encoder.AppendByte(byte(packet.MessageTypeId))
	}
	if packet.Fmt == 0 {
		//有流id
		encoder.EncodeInt32LittleEndian(int32(packet.MessageStreamId))
	}
	//时间扩展
	if packet.TimeStamp >= 0xffffff {
		encoder.EncodeInt32(int32(packet.TimeStamp))
	}
	//消息
	var bodySended uint32
	bodySended = 0
	sendSize := packet.MessageLength
	if sendSize > this.ChunkSize {
		sendSize = this.ChunkSize
	}
	encoder.AppendByteArray(packet.Body[:sendSize])
	buf, err := encoder.GetData()
	if err != nil {
		return
	}

	_, err = wssAPI.TcpWrite(this.Conn, buf)
	if err != nil {
		return
	}
	bodySended += sendSize
	//剩下的消息
	packet.Fmt = 3
	for bodySended < packet.MessageLength {
		encoder3 := &AMF0Encoder{}
		encoder3.Init()
		if packet.ChunkStreamID < 63 {
			tmp8 = (packet.Fmt << 6) | byte(packet.ChunkStreamID)
			encoder3.AppendByte(tmp8)
		} else if packet.ChunkStreamID < 319 {
			tmp8 = (packet.Fmt << 6)
			encoder3.AppendByte(tmp8)
			encoder3.AppendByte(byte(packet.ChunkStreamID - 64))
		} else if packet.ChunkStreamID < 65599 {
			tmp8 = (packet.Fmt << 6) | byte(1)
			encoder3.AppendByte(tmp8)
			encoder3.AppendByte(byte(packet.ChunkStreamID-64) & 0xff)
			encoder3.AppendByte(byte(packet.ChunkStreamID-64) >> 8)
		}
		sendSize = packet.MessageLength - bodySended
		if sendSize > this.ChunkSize {
			sendSize = this.ChunkSize
		}
		encoder3.AppendByteArray(packet.Body[bodySended : bodySended+sendSize])
		buf, err := encoder3.GetData()
		if err != nil {
			return err
		}
		_, err = wssAPI.TcpWrite(this.Conn, buf)
		if err != nil {
			return err
		}
		bodySended += sendSize
	}
	if RTMP_PACKET_TYPE_INVOKE == packet.MessageTypeId && queue {
		cmdName, err := AMF0DecodeString(packet.Body[1:])
		if err != nil {
			return err
		}
		transactionId, err := AMF0DecodeNumber(packet.Body[4+len(cmdName):])
		if err != nil {
			return err
		}
		this.methodCache[int32(transactionId)] = cmdName
	}
	return
}
