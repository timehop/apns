package apns

import (
	"crypto/tls"
	"io"
	"net"
	"strings"
	"time"
)

const (
	ProductionGateway = "gateway.push.apple.com:2195"
	SandboxGateway    = "gateway.sandbox.push.apple.com:2195"

	ProductionFeedbackGateway = "feedback.push.apple.com:2196"
	SandboxFeedbackGateway    = "feedback.sandbox.push.apple.com:2196"
)

// Conn is a wrapper for the actual TLS connections made to Apple
type Conn interface {
	io.ReadWriteCloser

	Connect() error
	ReadWithTimeout(p []byte, deadline time.Time) (int, error)
}

type conn struct {
	netConn net.Conn
	tls     *tls.Config

	gateway   string
	connected bool
}

func NewConnWithCert(gw string, cert tls.Certificate) Conn {
	gatewayParts := strings.Split(gw, ":")
	tls := tls.Config{
		Certificates:       []tls.Certificate{cert},
		ServerName:         gatewayParts[0],
		InsecureSkipVerify: true,
	}

	return &conn{gateway: gw, tls: &tls}
}

// NewConnWithFiles creates a new Conn from certificate and key in the specified files
func NewConn(gw string, crt string, key string) (Conn, error) {
	cert, err := tls.X509KeyPair([]byte(crt), []byte(key))
	if err != nil {
		return &conn{}, err
	}

	return NewConnWithCert(gw, cert), nil
}

// NewConnWithFiles creates a new Conn from certificate and key in the specified files
func NewConnWithFiles(gw string, certFile string, keyFile string) (Conn, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return &conn{}, err
	}

	return NewConnWithCert(gw, cert), nil
}

// Connect actually creates the TLS connection
func (c *conn) Connect() error {
	// Make sure the existing connection is closed
	if c.netConn != nil {
		c.netConn.Close()
	}

	tlsConn, err := tls.Dial("tcp", c.gateway, c.tls)
	if err != nil {
		return err
	}

	c.netConn = tlsConn
	return nil
}

func (c *conn) Close() error {
	if c.netConn != nil {
		return c.netConn.Close()
	}

	return nil
}

// Read reads data from the connection
func (c *conn) Read(p []byte) (int, error) {
	i, err := c.netConn.Read(p)
	return i, err
}

// ReadWithTimeout reads data from the connection and returns an error
// after duration
func (c *conn) ReadWithTimeout(p []byte, deadline time.Time) (int, error) {
	c.netConn.SetReadDeadline(deadline)
	i, err := c.netConn.Read(p)
	return i, err
}

// Write writes data from the connection
func (c *conn) Write(p []byte) (int, error) {
	return c.netConn.Write(p)
}
