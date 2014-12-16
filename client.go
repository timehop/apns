package apns

import (
	"container/list"
	"crypto/tls"
	"io"
	"log"
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
		id:           1,
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
	err := c.Conn.Connect()
	if err != nil {
		return err
	}

	go c.runLoop()
	return nil
}

func (c *Client) Send(n Notification) error {
	if !c.connected {
		return ErrDisconnected
	}

	// Set identifier if not specified
	if n.Identifier == 0 {
		n.Identifier = c.nextID()
	} else if c.id < n.Identifier {
		c.setID(n.Identifier)
	}

	b, err := n.ToBinary()
	if err != nil {
		return err
	}

	c.notifs <- serialized{b: b, id: n.Identifier, n: &n}
	return nil
}

func (c *Client) setID(n uint32) {
	c.idm.Lock()
	defer c.idm.Unlock()

	c.id = n
}

func (c *Client) nextID() uint32 {
	c.idm.Lock()
	defer c.idm.Unlock()

	c.id++
	return c.id
}

func (c *Client) connected() {
	c.connm.Lock()
	defer c.connm.Unlock()

	c.connected = true
}

func (c *Client) disconnected() {
	c.connm.Lock()
	defer c.connm.Unlock()

	c.connected = false
}

func (c *Client) reportFailedPush(s serialized, err *Error) {
	select {
	case c.FailedNotifs <- NotificationResult{Notif: *s.n, Err: *err}:
	default:
	}
}

func (c *Client) requeue(cursor *list.Element) {
	// If `cursor` is not nil, this means there are notifications that
	// need to be delivered (or redelivered)
	for ; cursor != nil; cursor = cursor.Next() {
		if n, ok := cursor.Value.(serialized); ok {
			go func() { c.notifs <- n }()
		}
	}
}

func (c *Client) handleError(err *Error, buffer *buffer) *list.Element {
	cursor := buffer.Back()

	for cursor != nil {
		// Get serialized notification
		n, _ := cursor.Value.(serialized)

		// If the notification, move cursor after the trouble notification
		if n.id == err.Identifier {
			go c.reportFailedPush(n, err)

			next := cursor.Next()

			buffer.Remove(cursor)
			return next
		}

		cursor = cursor.Prev()
	}

	return cursor
}

func (c *Client) runLoop() {
	sent := newBuffer(50)
	cursor := sent.Front()

	// APNS connection
	for {
		// Start reading errors from APNS
		errs := readErrs(c.Conn)

		c.requeue(cursor)

		// Connection open, listen for notifs and errors
		for {
			var err error
			var n serialized

			// Check for notifications or errors. There is a chance we'll send notifications
			// if we already have an error since `select` will "pseudorandomly" choose a
			// ready channels. It turns out to be fine because the connection will already
			// be closed and it'll requeue. We could check before we get to this select
			// block, but it doesn't seem worth the extra code and complexity.
			select {
			case err = <-errs:
			case n = <-c.notifs:
			}

			// If there is an error we understand, find the notification that failed,
			// move the cursor right after it.
			if nErr, ok := err.(*Error); ok {
				cursor = c.handleError(nErr, sent)
				break
			}

			if err != nil {
				break
			}

			// Add to list
			cursor = sent.Add(n)

			_, err = c.Conn.Write(n.b)

			if err == io.EOF {
				log.Println("EOF trying to write notification")
				c.connected = false
				return
			}

			if err != nil {
				log.Println("err writing to apns", err.Error())
				break
			}

			cursor = cursor.Next()
		}
	}
}

func readErrs(c *Conn) chan error {
	errs := make(chan error)

	go func() {
		p := make([]byte, 6, 6)
		_, err := c.Read(p)
		if err != nil {
			errs <- err
			return
		}

		e := NewError(p)
		errs <- &e
	}()

	return errs
}
