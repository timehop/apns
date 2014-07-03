package apns

import (
	"bytes"
	"encoding/binary"
	"time"
)

type Feedback struct {
	Conn *Conn
}

type FeedbackTuple struct {
	Timestamp   time.Time
	TokenLength uint16
	DeviceToken string
}

func feedbackTupleFromBytes(b []byte) FeedbackTuple {
	var ts uint32
	var tokLen uint16
	tok := make([]byte, 32)

	r := bytes.NewReader(b)

	binary.Read(r, binary.BigEndian, &ts)
	binary.Read(r, binary.BigEndian, &tokLen)
	binary.Read(r, binary.BigEndian, &tok)

	return FeedbackTuple{Timestamp: time.Unix(int64(ts), 0), TokenLength: tokLen, DeviceToken: string(tok)}
}

func NewFeedback(gw string, cert string, key string) (Feedback, error) {
	conn, err := NewConn(gw, cert, key)
	if err != nil {
		return Feedback{}, err
	}

	return Feedback{Conn: &conn}, nil
}

func NewFeedbackWithFiles(gw string, certFile string, keyFile string) (Feedback, error) {
	conn, err := NewConnWithFiles(gw, certFile, keyFile)
	if err != nil {
		return Feedback{}, err
	}

	return Feedback{Conn: &conn}, nil
}

func (f Feedback) Receive() <-chan FeedbackTuple {
	fc := make(chan FeedbackTuple)
	go f.receive(fc)
	return fc
}

func (f Feedback) receive(fc chan FeedbackTuple) {
	err := f.Conn.Connect()
	if err != nil {
		close(fc)
		return
	}
	defer f.Conn.Close()

	for {
		b := make([]byte, 38)

		f.Conn.NetConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		_, err := f.Conn.Read(b)
		if err != nil {
			close(fc)
			return
		}

		fc <- feedbackTupleFromBytes(b)
	}
}
