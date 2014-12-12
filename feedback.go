package apns

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
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
	r := bytes.NewReader(b)

	var ts uint32
	binary.Read(r, binary.BigEndian, &ts)

	var tokLen uint16
	binary.Read(r, binary.BigEndian, &tokLen)

	tok := make([]byte, tokLen)
	binary.Read(r, binary.BigEndian, &tok)

	return FeedbackTuple{
		Timestamp:   time.Unix(int64(ts), 0),
		TokenLength: tokLen,
		DeviceToken: hex.EncodeToString(tok),
	}
}

func NewFeedbackWithCert(gw string, cert tls.Certificate) Feedback {
	conn := NewConnWithCert(gw, cert)

	return Feedback{Conn: &conn}
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

// Receive returns a read only channel for APNs feedback. The returned channel
// will close when there is no more data to be read.
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
