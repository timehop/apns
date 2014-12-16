package apns

import (
	"container/list"
	"crypto/tls"
	"io"
	"sync"
)

type buffer struct {
	size int
	*list.List
}

func newBuffer(size int) *buffer {
	return &buffer{size, list.New()}
}

func (b *buffer) Add(v interface{}) *list.Element {
	e := b.PushBack(v)

	if b.Len() > b.size {
		b.Remove(b.Front())
	}

	return e
}

type serialized struct {
	id uint32
	b  []byte
	n  *Notification
}

type Client struct {
	Conn         *Conn
	FailedNotifs chan NotificationResult

	notifs chan serialized

	buffer *buffer
	cursor *list.Element

	id  uint32
	idm sync.Mutex

	connected bool
	connm     sync.Mutex
}

func newClientWithConn(gw string, conn Conn) Client {
	c := Client{
		Conn:         &conn,
		FailedNotifs: make(chan NotificationResult),
		notifs:       make(chan serialized),
		buffer:       newBuffer(50),
		cursor:       nil,
		id:           0,
		idm:          sync.Mutex{},
		connected:    false,
		connm:        sync.Mutex{},
	}

	return c
}

func NewClientWithCert(gw string, cert tls.Certificate) Client {
	conn := NewConnWithCert(gw, cert)

	return newClientWithConn(gw, conn)
}

func NewClient(gw string, cert string, key string) (Client, error) {
	conn, err := NewConn(gw, cert, key)
	if err != nil {
		return Client{}, err
	}

	return newClientWithConn(gw, conn), nil
}

func NewClientWithFiles(gw string, certFile string, keyFile string) (Client, error) {
	conn, err := NewConnWithFiles(gw, certFile, keyFile)
	if err != nil {
		return Client{}, err
	}

	return newClientWithConn(gw, conn), nil
}

func (c *Client) Connect() error {
	if err := c.Conn.Connect(); err != nil {
		return err
	}

	// On connect, requeue any notifications that were
	// sent after the error & disconnect.
	// http://redth.codes/the-problem-with-apples-push-notification-ser/
	if err := c.requeue(); err != nil {
		return err
	}

	// Kick off asynchronous error reading
	go c.readErrors()

	return nil
}

func (c *Client) Send(n Notification) error {
	if !c.connected {
		return ErrDisconnected
	}

	// Set identifier if not specified
	n.Identifier = c.determineIdentifier(n.Identifier)

	b, err := n.ToBinary()
	if err != nil {
		return err
	}

	// Add to list
	c.cursor = c.buffer.Add(n)

	_, err = c.Conn.Write(b)
	if err == io.EOF {
		c.connected = false
		return err
	}

	if err != nil {
		return err
	}

	c.cursor = c.cursor.Next()
	return nil
}

func (c *Client) determineIdentifier(n uint32) uint32 {
	c.idm.Lock()
	defer c.idm.Unlock()

	// If the id passed in is 0, that means it wasn't
	// set so get the next ID. Otherwise, set it to that
	// identifier.
	if n == 0 {
		c.id++
	} else {
		c.id = n
	}

	return c.id
}

func (c *Client) requeue() error {
	// If `cursor` is not nil, this means there are notifications that
	// need to be delivered (or redelivered)
	for ; c.cursor != nil; c.cursor = c.cursor.Next() {
		if s, ok := c.cursor.Value.(serialized); ok {
			if err := c.Send(*s.n); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) readErrors() {
	p := make([]byte, 6, 6)

	_, err := c.Conn.Read(p)
	// TODO(bw) not sure what to do here. It's unclear what errors
	// come out of this and how we handle it.
	if err != nil {
		return
	}

	e := NewError(p)
	cursor := c.buffer.Back()

	for cursor != nil {
		// Get serialized notification
		s, _ := cursor.Value.(serialized)

		// If the notification, move cursor after the trouble notification
		if s.id == e.Identifier {
			// Try to write - skip if no one is reading on the other side
			select {
			case c.FailedNotifs <- NotificationResult{Notif: *s.n, Err: e}:
			default:
			}

			c.cursor = cursor.Next()
			c.buffer.Remove(cursor)
		}

		cursor = cursor.Prev()
	}
}
