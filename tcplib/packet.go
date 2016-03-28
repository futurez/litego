package tcplib

import (
	"bytes"
	"encoding/binary"
)

const (
	LENGTH_SIZE = 4
	MSGID_SIZE  = 4
)

type Packet struct {
	MsgId uint32
	Data  []byte
}

func decodePacket(buff *[]byte) (uint32, []byte, bool) {
	if len(*buff) < LENGTH_SIZE {
		return 0, nil, false
	}

	//read size
	reader := bytes.NewReader((*buff)[:LENGTH_SIZE])
	var msgLen uint32
	binary.Read(reader, binary.LittleEndian, &msgLen)
	if uint32(len(*buff)) < msgLen+LENGTH_SIZE+MSGID_SIZE {
		return 0, nil, false
	}

	//read msgid
	reader = bytes.NewReader((*buff)[LENGTH_SIZE : LENGTH_SIZE+MSGID_SIZE])
	var msgId uint32
	binary.Read(reader, binary.LittleEndian, &msgId)

	//read data
	retBuf := (*buff)[LENGTH_SIZE+MSGID_SIZE : LENGTH_SIZE+MSGID_SIZE+msgLen]
	*buff = (*buff)[LENGTH_SIZE+MSGID_SIZE+msgLen:]

	return msgId, retBuf, true
}

func encodePacket(msgId uint32, buf *[]byte) []byte {
	msgLen := uint32(len(*buf))
	var w bytes.Buffer
	binary.Write(&w, binary.LittleEndian, &msgLen)
	binary.Write(&w, binary.LittleEndian, &msgId)
	binary.Write(&w, binary.LittleEndian, buf)
	return w.Bytes()
}
