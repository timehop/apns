package apns

import (
	"errors"
	"io"
	"sync"
)

// Session to Apple's Push Notification server.
type Session interface {
	Send(n Notification) error
	Connect() error
	RequeueableNotifications() []Notification
	Disconnect()
	Disconnected() bool
}

type sessionState int

const (
	sessionStateNew sessionState = 1 << iota
	sessionStateConnected
	sessionStateDisconnected
)

type session struct {
	b *buffer

	conn  Conn
	connm sync.Mutex

	st  sessionState
	stm sync.Mutex

	err Error
}

// NewSession creates a new session.
func NewSession(conn Conn) Session {
	return &session{
		st:    sessionStateNew,
		stm:   sync.Mutex{},
		conn:  conn,
		connm: sync.Mutex{},
		b:     newBuffer(50),
	}
}

// Connect session to gateway.
func (s *session) Connect() error {
	if s.isNew() {
		return errors.New("can't connect unless the session is new")
	}

	go s.readErrors()
	return nil
}

func (s *session) isNew() bool {
	s.stm.Lock()
	defer s.stm.Unlock()

	return s.st != sessionStateNew
}

// Disconnected indicates whether session is disconnected.
func (s *session) Disconnected() bool {
	s.stm.Lock()
	defer s.stm.Unlock()

	return s.st == sessionStateDisconnected
}

// Connected indicates whether session is connected.
func (s *session) Connected() bool {
	s.stm.Lock()
	defer s.stm.Unlock()

	return s.st == sessionStateConnected
}

// Send notification to gateway.
func (s *session) Send(n Notification) error {
	// If disconnected, error out
	if !s.Connected() {
		return errors.New("not connected")
	}

	// Serialize
	b, err := n.ToBinary()
	if err != nil {
		return err
	}

	// Add to buffer
	s.b.Add(n)

	// Send synchronously
	return s.write(b)
}

func (s *session) write(b []byte) error {
	s.connm.Lock()
	defer s.connm.Unlock()

	_, err := s.conn.Write(b)
	if err == io.EOF {
		s.Disconnect()
		return err
	}

	return err
}

// Disconnect from gateway.
func (s *session) Disconnect() {
	s.transitionState(sessionStateDisconnected)
}

func (s *session) transitionState(st sessionState) {
	s.stm.Lock()
	defer s.stm.Unlock()

	s.st = st
}

func (s *session) readErrors() {
	p := make([]byte, 6, 6)

	_, err := s.conn.Read(p)
	// TODO(bw) not sure what to do here. It's unclear what errors
	// come out of this and how we handle it.
	if err != nil {
		return
	}

	s.Disconnect()

	s.err = NewError(p)
}
