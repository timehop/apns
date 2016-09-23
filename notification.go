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
	// PriorityImmediate for time sensitive notifications (not for silent push messages).
	PriorityImmediate = 10
	// PriorityPowerConserve for notifications that are less time sensitive.
	PriorityPowerConserve = 5
)

const (
	validDeviceTokenLength = 64
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

// Alert to send.
type Alert struct {
	Body         string   `json:"body,omitempty"`
	Title        string   `json:"title,omitempty"`
	Action       string   `json:"action,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`

	// Do not add fields without updating the implementation of isZero.
}

// isSimple alerts only contain a Body.
func (a *Alert) isSimple() bool {
	return len(a.Title) == 0 && len(a.Action) == 0 && len(a.LocKey) == 0 && len(a.LocArgs) == 0 && len(a.ActionLocKey) == 0 && len(a.LaunchImage) == 0
}

func (a *Alert) isZero() bool {
	return a.isSimple() && len(a.Body) == 0
}

// APS is the Apple-reserved aps namespace in a push notification.
type APS struct {
	Alert            Alert
	Badge            BadgeNumber
	Sound            string
	ContentAvailable int
	URLArgs          []string
	Category         string // requires iOS 8+
	AccountId        string // for email push notifications
}

// MarshalJSON implements the json.Marshaler interface.
func (aps APS) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	if !aps.Alert.isZero() {
		if aps.Alert.isSimple() {
			data["alert"] = aps.Alert.Body
		} else {
			data["alert"] = aps.Alert
		}
	}
	if aps.Badge.IsSet {
		data["badge"] = aps.Badge
	}
	if aps.Sound != "" {
		data["sound"] = aps.Sound
	}
	if aps.ContentAvailable != 0 {
		data["content-available"] = aps.ContentAvailable
	}
	if aps.Category != "" {
		data["category"] = aps.Category
	}
	if aps.URLArgs != nil && len(aps.URLArgs) != 0 {
		data["url-args"] = aps.URLArgs
	}
	if aps.AccountId != "" {
		data["account-id"] = aps.AccountId
	}

	return json.Marshal(data)
}

// Payload to send Apple.
type Payload struct {
	APS APS
	// MDM for mobile device management
	MDM          string
	customValues map[string]interface{}
}

// MarshalJSON implements the json.Marshaler interface.
func (p *Payload) MarshalJSON() ([]byte, error) {
	if len(p.MDM) != 0 {
		p.customValues["mdm"] = p.MDM
	} else {
		p.customValues["aps"] = p.APS
	}

	return json.Marshal(p.customValues)
}

// SetCustomValue sets a custom payload value.
func (p *Payload) SetCustomValue(key string, value interface{}) error {
	if key == "aps" {
		return errors.New("cannot assign a custom APS value in payload")
	}

	p.customValues[key] = value

	return nil
}

// Notification contains the payload.
type Notification struct {
	ID          string
	DeviceToken string
	Identifier  uint32
	Expiration  *time.Time
	Priority    int
	Payload     *Payload
}

// NewNotification creates a new notification.
func NewNotification() Notification {
	return Notification{Payload: NewPayload()}
}

// NewPayload creates a new payload.
func NewPayload() *Payload {
	return &Payload{customValues: map[string]interface{}{}}
}

// ToBinary encodes a notification to send it.
func (n Notification) ToBinary() ([]byte, error) {
	b := []byte{}

	if len(n.DeviceToken) != validDeviceTokenLength {
		return b, errors.New(ErrInvalidToken)
	}

	binTok, err := hex.DecodeString(n.DeviceToken)
	if err != nil {
		return b, fmt.Errorf("convert token to hex error: %s", err)
	}

	j, err := json.Marshal(n.Payload)
	if err != nil {
		return b, fmt.Errorf("json marshal error: %s", err)
	}

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
