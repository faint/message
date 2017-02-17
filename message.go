package Server

import (
	"bytes"
	"encoding/binary"
)

// const ...
const (
	SizeofType = 4 // Sizeof Message Type
	SizeofSize = 4 // Sizeof Message Size
)

// Message used for communication between servers and clients
type Message struct {
	Type    int    // 消息头
	Size    int    // 消息大小
	Content []byte // 消息内容
}

// Unpack message from the bytes conn.Read()
func Unpack(b []byte) (Message, error) {
	m := Message{}
	buf := bytes.NewBuffer(b)

	// read type
	mType := buf.Next(SizeofType)
	bufType := bytes.NewBuffer(mType)
	binary.Read(bufType, binary.LittleEndian, &m.Type)
	// read size

	// read content

	return m, nil
}

func toInt(buffer *bytes.Buffer, size int) int {
	var n int
	binary.Read(bytes.NewBuffer(buffer.Next(size)), binary.LittleEndian, n)
	return n
}
