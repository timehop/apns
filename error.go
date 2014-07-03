package apns

import (
	"bytes"
	"encoding/binary"
)

const (
	// Error strings based on the codes specified here:
	// https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW12
	ErrProcessing         = "Processing error"
	ErrMissingDeviceToken = "Missing device token"
	ErrMissingTopic       = "Missing topic"
	ErrMissingPayload     = "Missing payload"
	ErrInvalidTokenSize   = "Invalid token size"
	ErrInvalidTopicSize   = "Invalid topic size"
	ErrInvalidPayloadSize = "Invalid payload size"
	ErrInvalidToken       = "Invalid token"
	ErrShutdown           = "Shutdown"
	ErrUnknown            = "None (unknown)"
)

var errorMapping = map[uint8]string{
	1:   ErrProcessing,
	2:   ErrMissingDeviceToken,
	3:   ErrMissingTopic,
	4:   ErrMissingPayload,
	5:   ErrInvalidTokenSize,
	6:   ErrInvalidTopicSize,
	7:   ErrInvalidPayloadSize,
	8:   ErrInvalidToken,
	10:  ErrShutdown,
	255: ErrUnknown,
}

type Error struct {
	Command    uint8
	Status     uint8
	Identifier uint32
	ErrStr     string
}

func NewError(p []byte) Error {
	if len(p) != 1+1+4 {
		return Error{ErrStr: ErrUnknown}
	}

	r := bytes.NewBuffer(p)
	e := Error{}

	binary.Read(r, binary.BigEndian, &e.Command)
	binary.Read(r, binary.BigEndian, &e.Status)
	binary.Read(r, binary.BigEndian, &e.Identifier)

	var ok bool
	if e.ErrStr, ok = errorMapping[e.Status]; !ok {
		e.ErrStr = ErrUnknown
	}

	return e
}

func (e *Error) Error() string {
	return e.ErrStr
}
