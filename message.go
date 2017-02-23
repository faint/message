package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

// const ...
const (
	SizeofType = 4    // Sizeof Message Type
	SizeofSize = 4    // Sizeof Message Size
	MaxBuffer  = 1024 // sizeofSize = 4(means int32), so max is 2,147,483,647
	SizeofHead = SizeofType + SizeofSize
	//
	ExpiredReadConnTime = 60 * 10 // time.Second
	RetryReadConnTime   = 100     // time.Microsecond
)

// Message used for communication between servers and clients
type Message struct {
	Type    int    // 消息头
	Size    int    // 消息大小
	Content []byte // 消息内容
}

// Unpack message from the bytes conn.Read()
func Unpack(content []byte) (Message, error) {
	m := Message{}
	b := bytes.NewBuffer(content)

	m.Type = readIntInBuffer(b, SizeofType)
	m.Size = readIntInBuffer(b, SizeofSize)

	if m.Size < 0 || m.Size > MaxBuffer { // illegal size
		return m, errors.New("BufferOverflow")
	}

	if m.Size == 0 { // head only , it's allowed
		return m, nil
	}

	m.Content = b.Bytes()
	return m, nil
}

// Pack message，返回[]byte
func Pack(messageType int, messageContent []byte) ([]byte, error) {
	b := new(bytes.Buffer)

	writeBuffer(b, int32(messageType))

	sizeofMessage := len(messageContent)
	if sizeofMessage > MaxBuffer {
		return nil, errors.New("OVER_MAX_BUFFER")
	}
	writeBuffer(b, int32(SizeofHead+sizeofMessage))

	writeBuffer(b, messageContent)
	return b.Bytes(), nil
}

func readIntInBuffer(buffer *bytes.Buffer, size int) int {
	var n int32
	binary.Read(bytes.NewBuffer(buffer.Next(size)), binary.LittleEndian, &n)
	return int(n)
}

func writeBuffer(buffer *bytes.Buffer, content interface{}) {
	binary.Write(buffer, binary.LittleEndian, content)
}

func read(conn net.Conn, size int) (Message, error) {
	b, err := readConn(conn, SizeofHead) // read head
	if err != nil {
		// todo
	}

	m, err := Unpack(b)
	if err != nil {
		// todo
	}
	if m.Size == 0 {
		return m, nil
	}

	c, err := readConn(conn, m.Size)
	if err != nil {
		// todo
	}
	m.Content = c
	return m, nil
}

func readConn(conn net.Conn, size int) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second * ExpiredReadConnTime))

	var b []byte
	unreadSize := size
	for unreadSize > 0 {
		tempBuf := make([]byte, unreadSize)
		n, err := conn.Read(tempBuf)
		if err != nil && err != io.EOF { // read error
			return nil, err
		}

		b = append(b[:size-unreadSize], tempBuf[:n]...)
		unreadSize -= n

		if unreadSize > 0 {
			time.Sleep(time.Microsecond * RetryReadConnTime)
		}
	}

	return b, nil
}
