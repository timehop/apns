package apns

import (
	"bytes"
	"encoding/binary"
)

const (
	// Error strings based on the codes specified here:
	// https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW12

	// ErrProcessing (1)
	ErrProcessing = "Processing error"
	// ErrMissingDeviceToken (2)
	ErrMissingDeviceToken = "Missing device token"
	// ErrMissingTopic (3)
	ErrMissingTopic = "Missing topic"
	// ErrMissingPayload (4) when no payload.
	ErrMissingPayload = "Missing payload"
	// ErrInvalidTokenSize (5) for a device token that is the wrong size.
	ErrInvalidTokenSize = "Invalid token size"
	// ErrInvalidTopicSize (6)
	ErrInvalidTopicSize = "Invalid topic size"
	// ErrInvalidPayloadSize (7) for a payload over 2 KB.
	ErrInvalidPayloadSize = "Invalid payload size"
	// ErrInvalidToken (8) such as a production device token used with the sandbox gateway.
	ErrInvalidToken = "Invalid token"
	// ErrShutdown (10) closed connection to perform maintenance. Open a new connection.
	ErrShutdown = "Shutdown"
	// ErrUnknown (255)
	ErrUnknown = "None (unknown)"
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

// Error captures the details of an error read from Apple's Push Notification server.
type Error struct {
	Command    uint8
	Status     uint8
	Identifier uint32
	ErrStr     string
}

// NewError parses an error from Apple.
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
