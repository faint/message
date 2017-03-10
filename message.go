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
	MaxBuffer         = 1024 // sizeofSize = 4(means int32), so max is 2,147,483,647.
	SizeofType        = 4    // Sizeof Message Type.
	SizeofSize        = 4    // Sizeof Message Size.
	SizeofTypeSize    = SizeofType + SizeofSize
	ExpiredConnWait   = 60 * 5 // *time.Second for read & write deadline.
	RetryReadConnTime = 100    // time.Microsecond.
)

// Message used for communication between servers and clients.
type Message struct {
	Type    int
	Size    int
	Content []byte
}

// Read use ReadConn to read and return Message.
func Read(conn net.Conn) (Message, error) {
	// read Message.Type + Message.Size
	b, err := ReadConn(conn, SizeofTypeSize)
	if err != nil {
		return Message{}, err
	}
	// fmt.Println("ReadConn:", b)
	m, err := Unpack(b) // Unpack head for content size.
	if err != nil {
		return m, err
	}
	// fmt.Println("Unpack:", m)
	if m.Size == 0 { // Message only has head, return.
		return m, nil
	}
	// read Message.Content
	c, err := ReadConn(conn, m.Size) // read content from conn.
	if err != nil {
		return m, err
	}
	// fmt.Println("ReadConn again:", c)
	m.Content = c
	// fmt.Println("return:", m)
	return m, nil
}

// Write use WriteConn to write the Message.
func Write(conn net.Conn, m Message) error {
	b, err := Pack(m.Type, m.Content)
	if err != nil {
		return err
	}
	err = WriteConn(conn, b)
	if err != nil {
		return err
	}
	return nil
}

// Unpack message from the bytes conn.Read().
func Unpack(content []byte) (Message, error) {
	m := Message{}
	b := bytes.NewBuffer(content)
	m.Type = readIntInBuffer(b, SizeofType)
	// fmt.Println("Type:", m.Type)
	m.Size = readIntInBuffer(b, SizeofSize)
	// fmt.Println("Size:", m.Size)
	if m.Size < 0 || m.Size > MaxBuffer { // illegal size.
		return m, errors.New("BufferOverflow")
	}
	if m.Size == 0 { // head only , it's allowed.
		return m, nil
	}
	m.Content = b.Bytes()
	// fmt.Println("Content:", m.Content)
	return m, nil
}

// Pack message，返回[]byte.
func Pack(messageType int, messageContent []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, int32(messageType))
	// fmt.Println("type:", b.Bytes())
	sizeofMessage := len(messageContent)
	if sizeofMessage > MaxBuffer {
		return nil, errors.New("OVER_MAX_BUFFER")
	}
	binary.Write(b, binary.LittleEndian, int32(sizeofMessage))
	// fmt.Println("+size:", b.Bytes())
	binary.Write(b, binary.LittleEndian, messageContent)
	// fmt.Println("+content:", b.Bytes())
	return b.Bytes(), nil
}

// ReadConn use conn.Read to read size []byte.
func ReadConn(conn net.Conn, size int) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second * ExpiredConnWait))
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

// WriteConn use conn.Write to write message.
func WriteConn(conn net.Conn, b []byte) error {
	conn.SetWriteDeadline(time.Now().Add(time.Second * ExpiredConnWait))
	_, err := conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// readIntInBuffer use int32 to read int in bytes.Buffer, and return int.
// use int32 because of the Message.Size is saved by 4byte, which means int32.
func readIntInBuffer(buffer *bytes.Buffer, size int) int {
	var n int32
	binary.Read(bytes.NewBuffer(buffer.Next(size)), binary.LittleEndian, &n)
	return int(n)
}
