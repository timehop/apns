package apns

import (
	"container/list"
	"crypto/tls"
	"io"
	"log"
	"sync/atomic"
	"time"
)

const (
	connectionMaxWaitSeconds = 300
)

var (
	atomicId int32 = 0
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

type Client struct {
	Conn         *Conn
	FailedNotifs chan NotificationResult
	notifs       chan Notification
	id           uint32
	clientId     int32
}

func newClientWithConn(gw string, conn Conn) Client {
	c := Client{
		Conn:         &conn,
		FailedNotifs: make(chan NotificationResult),
		notifs:       make(chan Notification),
		id:           uint32(1),
		clientId:     atomic.AddInt32(&atomicId, 1),
	}

	go c.runLoop()

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

func (c *Client) GetId() int32 {
	return c.clientId
}

func (c *Client) Send(n Notification) error {
	c.notifs <- n
	return nil
}

func (c *Client) reportFailedPush(v interface{}, err *Error) {
	failedNotif, ok := v.(Notification)
	if !ok || v == nil {
		return
	}

	select {
	case c.FailedNotifs <- NotificationResult{Notif: failedNotif, Err: *err}:
	default:
	}
}

func (c *Client) requeue(cursor *list.Element) {
	// If `cursor` is not nil, this means there are notifications that
	// need to be delivered (or redelivered)
	for ; cursor != nil; cursor = cursor.Next() {
		if n, ok := cursor.Value.(Notification); ok {
			go func() { c.notifs <- n }()
		}
	}
}

func (c *Client) handleError(err *Error, buffer *buffer) *list.Element {
	cursor := buffer.Back()

	for cursor != nil {
		// Get notification
		n, _ := cursor.Value.(Notification)

		// If the notification, move cursor after the trouble notification
		if n.Identifier == err.Identifier {
			go c.reportFailedPush(cursor.Value, err)

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
	numSent := uint64(0)

	// APNS connection
	for {
		err := c.Conn.Connect()
		if err != nil {
			time.Sleep(time.Second)
			continue
		} else {
			numSent = 0
		}

		// Start reading errors from APNS
		errs := readErrs(c.Conn)

		c.requeue(cursor)

		// Connection open, listen for notifs and errors
		for {
			var err error
			var n Notification

			select {
			case err = <-errs:
			case n = <-c.notifs:
				numSent++
			}

			// If there is an error we understand, find the notification that failed,
			// move the cursor right after it.
			if nErr, ok := err.(*Error); ok {
				log.Println("Known error:", err)
				cursor = c.handleError(nErr, sent)
				break
			}

			if err != nil {
				log.Println("Error on apns connection:", err)
				break
			}

			// Set identifier if not specified
			if n.Identifier == 0 {
				n.Identifier = c.id
				c.id++
			} else if c.id < n.Identifier {
				c.id = n.Identifier + 1
			}

			// Add to list
			cursor = sent.Add(n)

			b, err := n.ToBinary()
			if err != nil {
				log.Println("Error on converting notification to binary, error:", err)
				continue
			}

			log.Printf("Sending #%d notification (id: %s) in #%d connection\n", numSent, n.ID, c.clientId)

			written, err := c.Conn.Write(b)
			if err == io.EOF {
				log.Printf("EOF trying to write notification in #%d connection\n", c.clientId)
				break
			} else if err != nil {
				log.Printf("Error writing to apns %s in #%d connection", err.Error(), c.clientId)
				break
			} else if written < len(b) {
				log.Printf("Error: partial notification was written in #%d connection, notification id: %s\n",
					c.clientId, n.ID)
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
		n, err := c.Read(p)
		if n > 0 {
			e := NewError(p)
			errs <- &e
		} else if err != nil {
			errs <- err
		}
	}()

	return errs
}
