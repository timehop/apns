package apns

import (
	"crypto/tls"
	"sync"
	"time"
)

// Client creates a session with Apple and handles reconnection.
type Client struct {
	conn Conn

	sess  Session
	sessm sync.Mutex
}

func newClientWithConn(conn Conn) (Client, error) {
	c := Client{conn: conn}

	sess := newSession(conn)
	err := sess.Connect()
	if err != nil {
		return c, err
	}

	return Client{conn, sess, sync.Mutex{}}, nil
}

// NewClientWithCert creates a client of the Apple gateway given a certificate.
func NewClientWithCert(gw string, cert tls.Certificate) (Client, error) {
	conn := NewConnWithCert(gw, cert)
	return newClientWithConn(conn)
}

// NewClient is a helper that accepts a certificate/key pair.
func NewClient(gw string, cert string, key string) (Client, error) {
	conn, err := NewConn(gw, cert, key)
	if err != nil {
		return Client{}, err
	}

	return newClientWithConn(conn)
}

// NewClientWithFiles is a helper that loads a certificate/key from files.
func NewClientWithFiles(gw string, certFile string, keyFile string) (Client, error) {
	conn, err := NewConnWithFiles(gw, certFile, keyFile)
	if err != nil {
		return Client{}, err
	}

	return newClientWithConn(conn)
}

// Send a notification, handling disconnections.
func (c *Client) Send(n Notification) error {
	if c.sess.Disconnected() {
		c.reconnectAndRequeue()
	}

	return c.sess.Send(n)
}

func (c *Client) reconnectAndRequeue() {
	c.sessm.Lock()
	defer c.sessm.Unlock()

	// Pull off undelivered notifications
	notifs := c.sess.RequeueableNotifications()

	// Reconnect
	c.sess = nil

	for c.sess == nil {
		sess := newSession(c.conn)

		err := sess.Connect()
		if err != nil {
			// TODO retry policy
			// TODO connect error channel
			// Keep trying to connect
			time.Sleep(1 * time.Second)
			continue
		}

		c.sess = sess
	}

	for _, n := range notifs {
		// TODO handle error from sending
		c.sess.Send(n)
	}
}

var newSession = func(c Conn) Session {
	return NewSession(c)
}
