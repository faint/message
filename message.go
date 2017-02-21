package Server

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// const ...
const (
	SizeofType = 4    // Sizeof Message Type
	SizeofSize = 4    // Sizeof Message Size
	MaxBuffer  = 1024 // sizeofSize = 4(means int32), so max is 2 147 483 647
	SizeofHead = SizeofType + SizeofSize
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

	m.Type = read(buf, SizeofType)
	m.Size = read(buf, SizeofSize)
	if m.Size < 0 || m.Size > MaxBuffer {
		return m, errors.New("OVER_MAX_BUFFER")
	}
	mContent := buf.Bytes()
	rest := int(m.Size - SizeofHead)
	if rest > 0 {
		m.Content = mContent[:rest]
	}

	return m, nil
}

// Pack message，返回[]byte
func Pack(mType int, mContent []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	write(buf, int32(mType))
	sizeofMessage := len(mContent)
	if sizeofMessage > MaxBuffer {
		return nil, errors.New("OVER_MAX_BUFFER")
	}
	write(buf, int32(SizeofHead+sizeofMessage))
	write(buf, mContent)
	return buf.Bytes(), nil
}

func read(buffer *bytes.Buffer, size int) int {
	var n int32
	binary.Read(bytes.NewBuffer(buffer.Next(size)), binary.LittleEndian, &n)
	return int(n)
}

func write(buffer *bytes.Buffer, content interface{}) {
	binary.Write(buffer, binary.LittleEndian, content)
}
