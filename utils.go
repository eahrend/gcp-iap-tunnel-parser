package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// every binary struct pack that google does seems to be in big endian

/*
SUBPROTOCOL_TAG_CONNECT_SUCCESS_SID = 0x0001
SUBPROTOCOL_TAG_RECONNECT_SUCCESS_ACK = 0x0002
SUBPROTOCOL_TAG_DATA = 0x0004
SUBPROTOCOL_TAG_ACK = 0x0007

  return (struct.unpack(str('>H'), binary_data[:2])[0],
          binary_data[2:])
*/
//tag, bytes_left = utils.ExtractSubprotocolTag(binary_data)
func extractSubProtocolTag(data []byte) (uint16, []byte, error) {
	if len(data) < 2 {
		return 0, nil, fmt.Errorf("incomplete data")
	}
	i := binary.BigEndian.Uint16(data[:2])
	return i, data[2:], nil
}

func handleSubprotocolConnectSuccessSid(data []byte) ([]byte, []byte, error) {
	return extractSubprotocolConnectSuccessSid(data)
}

func extractSubprotocolConnectSuccessSid(data []byte) ([]byte, []byte, error) {
	nextBytes, binaryData, err := extractUnsignedInt32(data)
	if err != nil {
		return nil, nil, err
	}
	return extractBinaryArray(binaryData, int(nextBytes))
}

func extractUnsignedInt32(data []byte) (uint32, []byte, error) {
	if len(data) < 4 {
		return 0, nil, fmt.Errorf("incomplete data")
	}
	dataLength := binary.BigEndian.Uint32(data[:4])
	return dataLength, data[4:], nil
}

func extractBinaryArray(data []byte, dataLen int) ([]byte, []byte, error) {
	if len(data) < dataLen {
		return nil, nil, fmt.Errorf("incomplete data")
	}

	return data[:dataLen], data[dataLen:], nil
}

func handleSubprotocolData(data []byte) ([]byte, []byte, error) {
	nextBytes, binaryData, err := extractUnsignedInt32(data)
	if err != nil {
		return nil, nil, err
	}
	return extractBinaryArray(binaryData, int(nextBytes))
}

type AckFrame struct {
	Tag      uint16
	Received uint64
}

// Q is uint64
// H is uint16
func sendAck(bytesReceived int) ([]byte, error) {
	af := AckFrame{
		Tag:      7,
		Received: uint64(bytesReceived),
	}
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, af)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//CreateSubprotocolDataFrame
// I is uint32
type DataFrame struct {
	Tag uint16
	Len uint32
}

func createSubprotocolDataFrame(data []byte) []byte {
	df := DataFrame{
		Tag: 4,
		Len: uint32(len(data)),
	}
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, df)
	if err != nil {
		panic(err)
	}
	_, err = buf.Write(data)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}
