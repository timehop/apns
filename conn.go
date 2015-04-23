package apns

import (
	"crypto/tls"
	"io"
	"net"
	"strings"
	"time"
)

const (
	// ProductionGateway is the host for Apple Push Notification server.
	ProductionGateway = "gateway.push.apple.com:2195"
	// SandboxGateway is Apple's gateway for development.
	SandboxGateway = "gateway.sandbox.push.apple.com:2195"

	// ProductionFeedbackGateway is Apple's feedback service.
	ProductionFeedbackGateway = "feedback.push.apple.com:2196"
	// SandboxFeedbackGateway is Apple's feedback service for development.
	SandboxFeedbackGateway = "feedback.sandbox.push.apple.com:2196"
)

// Conn is a wrapper for the actual TLS connections made to Apple
type Conn interface {
	io.ReadWriteCloser

	Connect() error
	SetReadDeadline(deadline time.Time) error
}

type conn struct {
	netConn net.Conn
	tls     *tls.Config

	gateway   string
	connected bool
}

// NewConnWithCert creates a new connection from a certificate.
func NewConnWithCert(gw string, cert tls.Certificate) Conn {
	gatewayParts := strings.Split(gw, ":")
	tls := tls.Config{
		Certificates:       []tls.Certificate{cert},
		ServerName:         gatewayParts[0],
		InsecureSkipVerify: true,
	}

	return &conn{gateway: gw, tls: &tls}
}

// NewConn creates a new Conn from certificate and key pair.
func NewConn(gw string, crt string, key string) (Conn, error) {
	cert, err := tls.X509KeyPair([]byte(crt), []byte(key))
	if err != nil {
		return &conn{}, err
	}

	return NewConnWithCert(gw, cert), nil
}

// NewConnWithFiles creates a new Conn from certificate and key in the specified files.
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

// Close the connection.
func (c *conn) Close() error {
	if c.netConn != nil {
		return c.netConn.Close()
	}

	return nil
}

// Read reads data from the connection
func (c *conn) Read(p []byte) (int, error) {
	return c.netConn.Read(p)
}

// SetReadDeadline sets the read deadline on the underlying connection.
// A zero value for t means Read will not time out.
func (c *conn) SetReadDeadline(deadline time.Time) error {
	return c.netConn.SetReadDeadline(deadline)
}

// Write writes data from the connection
func (c *conn) Write(p []byte) (int, error) {
	return c.netConn.Write(p)
}
