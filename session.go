package apns

import (
	"container/list"
	"errors"
	"io"
	"sync"
)

// NotificationResult associates an error from Apple to a notification.
type NotificationResult struct {
	Notif Notification
	Err   Error
}

func (s NotificationResult) Error() string {
	return s.Err.Error()
}

// Session to Apple's Push Notification server.
type Session interface {
	Send(n Notification) error
	Connect() error
	RequeueableNotifications() []Notification
	Disconnect()
	Disconnected() bool
}

type buffer struct {
	size int
	m    sync.Mutex
	*list.List
}

func newBuffer(size int) *buffer {
	return &buffer{size, sync.Mutex{}, list.New()}
}

func (b *buffer) Add(v interface{}) *list.Element {
	b.m.Lock()
	defer b.m.Unlock()

	e := b.PushBack(v)

	if b.Len() > b.size {
		b.Remove(b.Front())
	}

	return e
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

	err NotificationResult
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
	return s.send(b)
}

func (s *session) send(b []byte) error {
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

// RequeueableNotifications returns good notifications sent after an error.
func (s *session) RequeueableNotifications() []Notification {
	notifs := []Notification{}

	// If still connected, return nothing
	if s.st != sessionStateDisconnected {
		return notifs
	}

	// Walk back to last known good notification and return the slice
	var e *list.Element
	for e = s.b.Front(); e != nil; e = e.Next() {
		if n, ok := e.Value.(Notification); ok && n.Identifier == s.err.Notif.Identifier {
			break
		}
	}

	// Start right after errored ID and get the rest of the list
	for e = e.Next(); e != nil; e = e.Next() {
		n, ok := e.Value.(Notification)
		if !ok {
			continue
		}

		notifs = append(notifs, n)
	}

	return notifs
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

	e := NewError(p)

	for cursor := s.b.Back(); cursor != nil; cursor = cursor.Prev() {
		// Get serialized notification
		n, _ := cursor.Value.(Notification)

		// If the notification, move cursor after the trouble notification
		if n.Identifier == e.Identifier {
			s.err = NotificationResult{n, e}
		}
	}
}
