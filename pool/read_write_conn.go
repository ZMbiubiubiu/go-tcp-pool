package pool

import (
	"encoding/binary"
	"io"
)

const prefixSize = 4

type onlyReader struct {
	io.Reader
}

// createTcpBuffer() implements the TCP protocol used in this application
// A stream of TCP data to be sent over has two parts: a prefix and the actual data itself
// The prefix is a fixed length byte that states how much data is being transferred over
func createTcpBuffer(data []byte) []byte {
	// Create a buffer with size enough to hold a prefix and actual data
	buf := make([]byte, prefixSize+len(data))

	// State the total number of bytes (including prefix) to be transferred over
	binary.BigEndian.PutUint32(buf[:prefixSize], uint32(prefixSize+len(data)))

	// Copy data into the remaining buffer
	copy(buf[prefixSize:], data[:])

	return buf
}

// ReadData reads the data from the underlying TCP connection
func (c *TcpConn) ReadData() ([]byte, error) {
	prefix := make([]byte, prefixSize)

	// Read the prefix, which contains the length of data expected
	_, err := io.ReadFull(c, prefix)
	if err != nil {
		return nil, err
	}

	totalDataLength := binary.BigEndian.Uint32(prefix[:])

	// Buffer to store the actual data
	data := make([]byte, totalDataLength-prefixSize)

	// Read actual data without prefix
	_, err = io.ReadFull(c, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
