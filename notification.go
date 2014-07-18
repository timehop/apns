package apns

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	PriorityImmediate     = 10
	PriorityPowerConserve = 5
)

const (
	commandID = 2

	// Items IDs
	deviceTokenItemID            = 1
	payloadItemID                = 2
	notificationIdentifierItemID = 3
	expirationDateItemID         = 4
	priorityItemID               = 5

	// Item lengths
	deviceTokenItemLength            = 32
	notificationIdentifierItemLength = 4
	expirationDateItemLength         = 4
	priorityItemLength               = 1
)

type NotificationResult struct {
	Notif Notification
	Err   Error
}

type Alert struct {
	Body         string   `json:"body,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`
}

type APS struct {
	Alert            Alert  `json:"alert,omitempty"`
	Badge            int    `json:"badge,omitempty"`
	Sound            string `json:"sound,omitempty"`
	ContentAvailable int    `json:"content-available,omitempty"`
}

type Payload struct {
	APS          APS
	customValues map[string]interface{}
}

type Notification struct {
	ID          string
	DeviceToken string
	Identifier  uint32
	Expiration  *time.Time
	Priority    int
	Payload     *Payload
}

func NewNotification() Notification {
	return Notification{Payload: NewPayload()}
}

func NewPayload() *Payload {
	return &Payload{customValues: map[string]interface{}{}}
}

func (p *Payload) SetCustomValue(key string, value interface{}) error {
	if key == "aps" {
		return errors.New("cannot assign a custom APS value in payload")
	}

	p.customValues[key] = value

	return nil
}

func (p *Payload) MarshalJSON() ([]byte, error) {
	p.customValues["aps"] = p.APS

	return json.Marshal(p.customValues)
}

func (n Notification) ToBinary() ([]byte, error) {
	b := []byte{}

	binTok, err := hex.DecodeString(n.DeviceToken)
	if err != nil {
		return b, fmt.Errorf("convert token to hex error: %s", err)
	}

	j, _ := json.Marshal(n.Payload)

	buf := bytes.NewBuffer(b)

	// Token
	binary.Write(buf, binary.BigEndian, uint8(deviceTokenItemID))
	binary.Write(buf, binary.BigEndian, uint16(deviceTokenItemLength))
	binary.Write(buf, binary.BigEndian, binTok)

	// Payload
	binary.Write(buf, binary.BigEndian, uint8(payloadItemID))
	binary.Write(buf, binary.BigEndian, uint16(len(j)))
	binary.Write(buf, binary.BigEndian, j)

	// Identifier
	binary.Write(buf, binary.BigEndian, uint8(notificationIdentifierItemID))
	binary.Write(buf, binary.BigEndian, uint16(notificationIdentifierItemLength))
	binary.Write(buf, binary.BigEndian, uint32(n.Identifier))

	// Expiry
	binary.Write(buf, binary.BigEndian, uint8(expirationDateItemID))
	binary.Write(buf, binary.BigEndian, uint16(expirationDateItemLength))
	if n.Expiration == nil {
		binary.Write(buf, binary.BigEndian, uint32(0))
	} else {
		binary.Write(buf, binary.BigEndian, uint32(n.Expiration.Unix()))
	}

	// Priority
	binary.Write(buf, binary.BigEndian, uint8(priorityItemID))
	binary.Write(buf, binary.BigEndian, uint16(priorityItemLength))
	binary.Write(buf, binary.BigEndian, uint8(n.Priority))

	framebuf := bytes.NewBuffer([]byte{})
	binary.Write(framebuf, binary.BigEndian, uint8(commandID))
	binary.Write(framebuf, binary.BigEndian, uint32(buf.Len()))
	binary.Write(framebuf, binary.BigEndian, buf.Bytes())

	return framebuf.Bytes(), nil
}
