package main

import (
	"errors"
	"fmt"
)

type IAPMessageInterface interface {
	GetTag() MessageTag
	SetTag(MessageTag)
	GetSequenceNumber() uint64
	SetSequenceNumber(uint64)
	ToString() string
}

type IAPMessage struct {
	data           []byte
	tag            MessageTag
	sequenceNumber uint64
}

type MessageTag uint16

const (
	MessageTagUnused MessageTag = iota
	MessageConnectSuccessSid
	MessageReconnectSuccessAck
	MessageDeprecated
	MessageData
	MessageAckLatency
	MessageReplyLatency
	MessageAck
)

func NewIAPMessage(data []byte) *IAPMessage {
	iapmsg := &IAPMessage{
		data: data,
	}
	return iapmsg
}

func (msg *IAPMessage) AsConnectSIDMessage() *IAPSidMessage {
	iapsidmsg := NewIAPSIDMessage(msg.data)
	return iapsidmsg
}

func (msg *IAPMessage) AsDataMessage() *IAPDataMessage {
	iapdatamsg := NewIAPDataMessage(msg.data)
	return iapdatamsg
}

func (msg *IAPMessage) PeekMessageTag() MessageTag {
	return getTag(msg.data, 0)
}

// GetMessageFromTag probably doesn't work as nicely as I'd like it to.
// the client calling it will need to figure out what to do.
func (msg *IAPMessage) GetMessageFromTag() (IAPMessageInterface, error) {
	tag := msg.PeekMessageTag()
	switch msgTag := tag; {
	case msgTag == MessageConnectSuccessSid:
		iapsidmsg := NewIAPSIDMessage(msg.data)
		return iapsidmsg, nil
	}
	return nil, errors.New(fmt.Sprintf("failed to get message tag: %d", tag))
}

func (msg *IAPMessage) ToString() string {
	return "not implemented on base class"
}

func (msg *IAPMessage) GetTag() MessageTag {
	return getTag(msg.data, 0)
}

func (msg *IAPMessage) SetTag(tag MessageTag) {
	msg.tag = tag
}

func (msg *IAPMessage) GetSequenceNumber() uint64 {
	return msg.sequenceNumber
}

func (msg *IAPMessage) SetSequenceNumber(sequenceNumber uint64) {
	msg.sequenceNumber = sequenceNumber
}

func getTag(data []byte, offset int) MessageTag {
	tag := decodeUint16(data, offset)
	return MessageTag(tag)
}

type IAPDataMessage struct {
	dataOffset     uint32
	maxTotalLength uint32
	maxDataLength  uint32
	data           []byte
	dataLength     uint32
	sequenceNumber uint64
	tag            MessageTag
}

func NewIAPDataMessage(data []byte) *IAPDataMessage {
	iapdm := &IAPDataMessage{
		dataOffset:     6,
		maxTotalLength: 65535,
		data:           data,
	}
	iapdm.maxDataLength = iapdm.maxTotalLength - iapdm.dataOffset
	return iapdm
}

func (msg *IAPDataMessage) GetTag() MessageTag {
	return getTag(msg.data, 0)
}
func (msg *IAPDataMessage) SetTag(tag MessageTag) {
	msg.tag = tag
}

func (msg *IAPDataMessage) GetSequenceNumber() uint64 {
	return msg.sequenceNumber
}

func (msg *IAPDataMessage) SetSequenceNumber(sequenceNumber uint64) {
	msg.sequenceNumber = sequenceNumber
}

func (msg *IAPDataMessage) GetExpectedAck() uint64 {
	return msg.sequenceNumber + uint64(msg.dataLength)
}

func (msg *IAPDataMessage) GetDataLength() uint32 {
	return decodeUint32(msg.data, 2)
}

func (msg *IAPDataMessage) SetDataLength(value uint32) error {
	if value < 0 || value > msg.maxDataLength {
		return errors.New("value out of range")
	}
	newData := encodeUint32(value, msg.data, 2)
	msg.dataLength = uint32(len(newData))
	return nil
}

func (msg *IAPDataMessage) GetBufferLength() int {
	return int(msg.dataOffset + msg.dataLength)
}

func (msg *IAPDataMessage) ToString() string {
	return fmt.Sprintf("Seq: %v, Len: %v, ExpAck: %v, Data: %s \r\n", msg.sequenceNumber, msg.dataLength, msg.GetExpectedAck(), string(msg.data))
}

type IAPSidMessage struct {
	dataOffset            uint32
	minimumExpectedLength uint32
	maxTotalLength        uint32
	maxDataLength         uint32
	data                  []byte
	sid                   string
	dataLength            uint32
	sequenceNumber        uint64
	tag                   MessageTag
}

func NewIAPSIDMessage(data []byte) *IAPSidMessage {
	iapsid := &IAPSidMessage{
		minimumExpectedLength: uint32(7),
		dataOffset:            uint32(6),
		data:                  data,
	}
	return iapsid
}
func (msg *IAPSidMessage) GetTag() MessageTag {
	return getTag(msg.data, 0)
}
func (msg *IAPSidMessage) SetTag(tag MessageTag) {
	msg.tag = tag
}

func (msg *IAPSidMessage) GetSequenceNumber() uint64 {
	return msg.sequenceNumber
}

func (msg *IAPSidMessage) SetSequenceNumber(sequenceNumber uint64) {
	msg.sequenceNumber = sequenceNumber
}

func (msg *IAPSidMessage) GetExpectedAck() uint64 {
	return msg.sequenceNumber + uint64(msg.dataLength)
}

func (msg *IAPSidMessage) GetDataLength() uint32 {
	return decodeUint32(msg.data, 2)
}

func (msg *IAPSidMessage) SetDataLength(value uint32) error {
	if value < 0 || value > msg.maxDataLength {
		return errors.New("value out of range")
	}
	newData := encodeUint32(value, msg.data, 2)
	msg.dataLength = uint32(len(newData))
	msg.data = newData
	return nil
}

func (msg *IAPSidMessage) GetBufferLength() int {
	return int(msg.dataOffset + msg.dataLength)
}

func (msg *IAPSidMessage) GetSID() string {
	newData := msg.data[msg.dataOffset:msg.GetDataLength()]
	return string(newData)
}

func (msg *IAPSidMessage) ToString() string {
	return fmt.Sprintf("SID: %v", msg.GetSID())
}

type IAPAckMessage struct {
	dataOffset     uint32
	maxTotalLength uint32
	expectedLength uint32
	maxDataLength  uint32
	ackOffset      uint32
	ack            uint64
	data           []byte
	dataLength     uint32
	sequenceNumber uint64
	tag            MessageTag
}

func NewIAPAckMessage(data []byte) *IAPAckMessage {
	iapack := &IAPAckMessage{
		data:           data,
		expectedLength: uint32(10),
		ackOffset:      uint32(2),
	}
	return iapack
}

func (msg *IAPAckMessage) GetTag() MessageTag {
	return getTag(msg.data, 0)
}
func (msg *IAPAckMessage) SetTag(tag MessageTag) {
	newData := encodeUint16(uint16(tag), msg.data, 0)
	msg.data = newData
	msg.tag = tag
}

func (msg *IAPAckMessage) GetSequenceNumber() uint64 {
	return msg.sequenceNumber
}

func (msg *IAPAckMessage) SetSequenceNumber(sequenceNumber uint64) {
	msg.sequenceNumber = sequenceNumber
}

func (msg *IAPAckMessage) GetAck() uint64 {
	return decodeUint64(msg.data, 2)
}

func (msg *IAPAckMessage) SetAck(value uint64) {
	newData := encodeUint64(value, msg.data, 2)
	msg.data = newData
	msg.ack = uint64(len(newData))
}

func (msg *IAPAckMessage) GetBufferLength() int {
	return 10
}

func (msg *IAPAckMessage) ToString() string {
	return fmt.Sprintf("Ack: %v", msg.ack)
}
