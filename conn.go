package apns

import (
	"crypto/tls"
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
type Conn struct {
	NetConn net.Conn
	Conf    *tls.Config
	timeout time.Duration

	gateway   string
	connected bool
}

func NewConnWithCertTimeout(gw string, cert tls.Certificate, timeout int) Conn {
	gatewayParts := strings.Split(gw, ":")
	conf := tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   gatewayParts[0],
	}

	return Conn{gateway: gw, Conf: &conf, timeout: time.Duration(timeout) * time.Second}
}

func NewConnWithCert(gw string, cert tls.Certificate) Conn {
	return NewConnWithCertTimeout(gw, cert, 0)
}

// NewConnWithFiles creates a new Conn from certificate and key in the specified files
func NewConn(gw string, crt string, key string) (Conn, error) {
	cert, err := tls.X509KeyPair([]byte(crt), []byte(key))
	if err != nil {
		return Conn{}, err
	}

	return NewConnWithCert(gw, cert), nil
}

// NewConnWithFiles creates a new Conn from certificate and key in the specified files
func NewConnWithFilesTimeout(gw string, certFile string, keyFile string, timeout int) (Conn, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return Conn{}, err
	}

	return NewConnWithCertTimeout(gw, cert, timeout), nil
}

// NewConnWithFiles creates a new Conn from certificate and key in the specified files
func NewConnWithFiles(gw string, certFile string, keyFile string) (Conn, error) {
	return NewConnWithFilesTimeout(gw, certFile, keyFile, 0)
}

// Connect actually creates the TLS connection
func (c *Conn) Connect() error {
	// Make sure the existing connection is closed
	if c.NetConn != nil {
		c.NetConn.Close()
	}

	var conn net.Conn
	var err error
	if c.timeout > 0 {
		conn, err = net.DialTimeout("tcp", c.gateway, c.timeout)
	} else {
		conn, err = net.Dial("tcp", c.gateway)
	}
	if err != nil {
		return err
	}

	tlsConn := tls.Client(conn, c.Conf)
	if c.timeout > 0 {
		tlsConn.SetDeadline(time.Now().Add(c.timeout * 3))
	}
	err = tlsConn.Handshake()
	if err != nil {
		return err
	}

	c.NetConn = tlsConn
	return nil
}

func (c *Conn) Close() error {
	if c.NetConn != nil {
		return c.NetConn.Close()
	}

	return nil
}

// Read reads data from the connection
func (c *Conn) Read(p []byte) (int, error) {
	i, err := c.NetConn.Read(p)
	return i, err
}

// Write writes data from the connection
func (c *Conn) Write(p []byte) (int, error) {
	if c.timeout > 0 {
		c.NetConn.SetWriteDeadline(time.Now().Add(c.timeout))
	}
	return c.NetConn.Write(p)
}
