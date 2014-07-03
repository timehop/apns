package apns_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/apns"
)

var DummyCert = `-----BEGIN CERTIFICATE-----
MIIC9TCCAd+gAwIBAgIQf3bEgFWUb+q6eK5ySkV/gjALBgkqhkiG9w0BAQUwEjEQ
MA4GA1UEChMHQWNtZSBDbzAeFw0xNDA2MzAwNDI5MDhaFw0xNTA2MzAwNDI5MDha
MBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDhAgWrrFZBtCfVEPg1tSIr9fuSUoeundb556IUr9uOmOHaYK7r3/I43acw
bVIfaenFxwUUf8YakQzTjOa5qSfK/Eylyw2ezBJtNUEqcHw0f+y66+jJbZa4clPa
tL6ezaMS/syXPpvNU8+16jdVdTJzqdBdSGAZMOCeumUWDNdlfBmHPVq1JMy0uGmO
XDoZK2Ir0/3LUfjk9R2wdm1VLrJAml7F0L0FhBHHXgHOSFM2ixjGflffaiuTCxhW
1z1NTo9XjWUQh2iM9Udf+xVnJLGLZ0EMFr2qihuK604Fp4SlNHEF+UWUn+j0PYo+
LbzM9oKJcdVD0XI36vrn3rGPHO9vAgMBAAGjSzBJMA4GA1UdDwEB/wQEAwIAoDAT
BgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMBQGA1UdEQQNMAuCCWxv
Y2FsaG9zdDALBgkqhkiG9w0BAQUDggEBAGJ/3I4KKlbEwLAC5ut4ZZ9V8WF4sHkI
Lj7e4vx2pPi6hf9miV1ff01NrpfUna7flwL9yD7Ybl7jRRIB4rIcKk+U5djGsT3H
ScGkbIMKrr08drWw1g4JU6PBH7xTfzGxNRERrnmrbJV0jCo9Tt8i53IpPtp6Z2Q1
8ydtPhU+Bpe2YoNr1w1fSV1JHXqjKV8RlGkCNSi4ozPOO8RbAYnBT3d9XSGoX//q
RGJUf3wC/rCxJkN63Moxuy3vxV2TmiqccHOrXJSJ8P/4PpPV/xuBk5k4HS1Nfmew
d9WHHn6bMJE9arVvWAiu9teCadVffuS2cl2cicN4XB6Ui0aDqhG2Exw=
-----END CERTIFICATE-----`

var DummyKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA4QIFq6xWQbQn1RD4NbUiK/X7klKHrp3W+eeiFK/bjpjh2mCu
69/yON2nMG1SH2npxccFFH/GGpEM04zmuaknyvxMpcsNnswSbTVBKnB8NH/suuvo
yW2WuHJT2rS+ns2jEv7Mlz6bzVPPteo3VXUyc6nQXUhgGTDgnrplFgzXZXwZhz1a
tSTMtLhpjlw6GStiK9P9y1H45PUdsHZtVS6yQJpexdC9BYQRx14BzkhTNosYxn5X
32orkwsYVtc9TU6PV41lEIdojPVHX/sVZySxi2dBDBa9qoobiutOBaeEpTRxBflF
lJ/o9D2KPi28zPaCiXHVQ9FyN+r6596xjxzvbwIDAQABAoIBAFzW+cIA5MJNdFX8
n32BlGzxHPEd7nAFHmuUwJKqkPwAZsg1NleK2qXOByr7IHRnvhZl7Nmtcu8JRHKR
Y63ddtbRTUrnQmJwL3YyEAZTzVvYILRrnGxoNFU8jw7hnvllPdEbow0QvzZ0S3Lz
BgvTxJJm0dt7fnNGcJftrsHvYHy1dptaR4hPv0xV5G7RPrbTl94llKfi745tp5Wd
xGpnjcBXoAnzCVRij1tHfSYubRJ2MJV0kzG3oVdRV2P/zWaout8BlhLCURv4sRUX
7FfCNa/z+G6AlROjCKJUP9YIUbxBEa/aP8YlSiyLRi1jFbMWcnKWQUdqS19m73Ap
a1LJFPECgYEA+Ve5DegcrWnUb2HsHD38HlmEg6S+/jg2P4TsuLZBtvO4/vzRx/qq
pwuuMm2CsvXr4nVmMEsMlSzYdsnaXIlWqyVDCOwIWR5VYT2GDWqQLaIXPlFaISzN
27tHd64KUtR1fMJUwQVK/MUORUbpYoAnSIil2SlYkWUhF024fNP8CxcCgYEA5wP4
HLiqU2rqe7vSAF/8fHwPleTzuCfMCVZm0aegUzQQQtklZoVE/BBwEGHdXflq1veq
pHeC8bNR4BF6ZgeSWgbLVF3msquy47QeNElHA2muJd3qmNWz4LXo1Pxb8KXcnXri
QZ+r3Y8obWTFQYq7gGQGPLXGTV3bhLGIyrT4lWkCgYAgZ2MYSJL5gmhmNT6fCPsr
4oxTI2Ti2uFJ7fdppd3ybcgb8zU8HPpyjRUNXqf+o/EM1B78pbQz6skS3vau0fZe
dZA5p5sKIeQMqBc0xSWJmKgWpDHnX9A8/yCxj/+tdgjytrqW/x4YrW9GV4nbEDaK
uZ98EmB9PLxJMAOKzW3S7wKBgQDD4PCy4b3CR2iVC9dva/P5VXQdo+knX884p6M8
58YgZofXNqnouN2aYRG0QlbiBMcbiRqOo6tK58JnnEpNUuQ8I4Cqg4hGPSHMwv/N
U8i70xLPltABUUpZIcVPOr92WBytBvHrtMiUb3tW7lf3T/vWTHmhZnvDQ+8LH0Ge
pz4T6QKBgQCoBJKOd781IQmT6i5hHSYJlsP6ymaaaQniJPVpnci/jf8+2QtponQY
scgnaBLBasLQ6GfKSRtcyidEi9wwxpVj0tw2p567jeNcIveD0TOYFf0RHEfrs+D4
VdRgai/v2NbFZLDnzeGVuYypXu6R78isJfHtz/a0aEave8yB3CRiDw==
-----END RSA PRIVATE KEY-----`

// To be able to run in parallel
var mockPort = 50000

// Mock Addr
type mockAddr struct {
}

func (m mockAddr) Network() string {
	return "localhost:56789"
}

func (m mockAddr) String() string {
	return "localhost:56789"
}

// Mock TLS connection
type mockTLSNetConn struct {
	bb  *bytes.Buffer
	err error
}

func (t mockTLSNetConn) Read(p []byte) (int, error) {
	r := bytes.NewReader(t.bb.Bytes())
	return r.Read(p)
}

func (t mockTLSNetConn) Write(p []byte) (int, error) {
	return t.bb.Write(p)
}

func (t mockTLSNetConn) Close() error {
	return t.err
}

func (m mockTLSNetConn) LocalAddr() net.Addr {
	return mockAddr{}
}

func (m mockTLSNetConn) RemoteAddr() net.Addr {
	return mockAddr{}
}

func (m mockTLSNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (m mockTLSNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m mockTLSNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type serverAction struct {
	action string
	data   []byte
	cb     func(s serverAction)
}

const (
	readAction  = "read"
	writeAction = "write"
	closeAction = "close"
)

type mockTLSServer struct {
	Port                   int
	Server                 net.Listener
	ConnectionActionGroups [][]serverAction
}

func (m *mockTLSServer) portStr() string {
	if m.Port == 0 {
		mockPort = mockPort + 1
		m.Port = mockPort
	}

	return fmt.Sprint(m.Port)
}

func (m *mockTLSServer) Address() string {
	return "localhost:" + m.portStr()
}

func (m *mockTLSServer) start() {
	cert, err := tls.X509KeyPair([]byte(DummyCert), []byte(DummyKey))
	if err != nil {
		log.Panic(err)
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAnyClientCert}

	m.Server, err = tls.Listen("tcp", "localhost:"+m.portStr(), &config)
	go func() {
		for i := 0; i < len(m.ConnectionActionGroups); i++ {
			g := m.ConnectionActionGroups[i]

			// Wait for a connection.
			conn, err := m.Server.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				} else {
					log.Fatal(err)
				}
			}
			// Handle the connection in a new goroutine.
			// The loop then returns to accepting, so that
			// multiple connections may be served concurrently.
			go func(c net.Conn) {
				for j := 0; j < len(g); j++ {
					a := g[j]
					switch a.action {
					case readAction:
						c.Read(a.data)
					case writeAction:
						c.Write(a.data)
					case closeAction:
						c.Close()

						if a.cb != nil {
							a.cb(a)
						}
						return
					}

					if a.cb != nil {
						a.cb(a)
					}
				}
			}(conn)
		}

		// No more connection action groups
	}()
}

func (m *mockTLSServer) stop() {
	if m.Server != nil {
		m.Server.Close()
	}
}

var withMockServer = func(as [][]serverAction, cb func(s *mockTLSServer)) {
	d := make(chan interface{})
	withMockServerAsync(as, d, func(s *mockTLSServer) {
		cb(s)
		close(d)
	})
}

var withMockServerAsync = func(as [][]serverAction, d chan interface{}, cb func(s *mockTLSServer)) {
	s := &mockTLSServer{}
	s.ConnectionActionGroups = as

	s.start()

	cb(s)

	<-d
	s.stop()
}

// Tests
var _ = Describe("Conn", func() {
	Describe(".NewConn", func() {
		Context("bad key/cert pair", func() {
			It("should return an error", func() {
				_, err := apns.NewConn(apns.SandboxGateway, "missing", "missing")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("valid key/cert pair", func() {
			It("should not return an error", func() {
				_, err := apns.NewConn(apns.SandboxGateway, DummyCert, DummyKey)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe(".NewConnWithFiles", func() {
		Context("missing files", func() {
			It("should return an error", func() {
				_, err := apns.NewConnWithFiles(apns.SandboxGateway, "missing.pem", "missing.pem")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with valid cert/key pair", func() {
			var certFile, keyFile *os.File
			var err error

			BeforeEach(func() {
				certFile, _ = ioutil.TempFile("", "cert.pem")
				certFile.Write([]byte(DummyCert))
				certFile.Close()

				keyFile, _ = ioutil.TempFile("", "key.pem")
				keyFile.Write([]byte(DummyKey))
				keyFile.Close()
			})

			AfterEach(func() {
				if certFile != nil {
					os.Remove(certFile.Name())
				}

				if keyFile != nil {
					os.Remove(keyFile.Name())
				}
			})

			It("should returning a connection", func() {
				_, err = apns.NewConnWithFiles(apns.SandboxGateway, certFile.Name(), keyFile.Name())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("#Connect()", func() {
		Context("server not up", func() {
			conn, _ := apns.NewConnWithFiles(apns.SandboxGateway, "missing.pem", "missing.pem")

			It("should return an error", func() {
				err := conn.Connect()
				Expect(err).NotTo(BeNil())
			})
		})

		Context("server up", func() {
			as := [][]serverAction{[]serverAction{serverAction{action: readAction, data: []byte{}}}}

			Context("with untrusted certs", func() {
				It("should return an error", func(d Done) {
					withMockServer(as, func(s *mockTLSServer) {
						conn, _ := apns.NewConn(s.Address(), DummyCert, DummyKey)
						err := conn.Connect()
						Expect(err).NotTo(BeNil())

						close(d)
					})
				})
			})

			Context("trusting the certs", func() {
				It("should not return an error", func(d Done) {
					withMockServer(as, func(s *mockTLSServer) {
						conn, _ := apns.NewConn(s.Address(), DummyCert, DummyKey)
						conn.Conf.InsecureSkipVerify = true

						err := conn.Connect()
						Expect(err).To(BeNil())

						close(d)
					})
				})
			})

			Context("with existing connection", func() {
				It("should not return an error", func(d Done) {
					as = [][]serverAction{
						[]serverAction{serverAction{action: readAction, data: []byte{}}},
						[]serverAction{serverAction{action: readAction, data: []byte{}}},
					}

					withMockServer(as, func(s *mockTLSServer) {
						conn, _ := apns.NewConn(s.Address(), DummyCert, DummyKey)
						conn.Conf.InsecureSkipVerify = true

						conn.Connect()

						err := conn.Connect()
						Expect(err).To(BeNil())

						close(d)
					})
				})
			})
		})
	})

	Describe("#Read", func() {
		rwc := mockTLSNetConn{bb: bytes.NewBuffer([]byte("hello!"))}

		pp := make([]byte, 6)
		bytes.NewReader(rwc.bb.Bytes()).Read(pp)

		conn, _ := apns.NewConn(apns.ProductionGateway, DummyCert, DummyKey)
		conn.NetConn = rwc

		It("should read out 'hello!'", func() {
			p := make([]byte, 6)
			conn.Read(p)

			Expect(p).To(Equal([]byte("hello!")))
		})
	})

	Describe("#Write", func() {
		rwc := mockTLSNetConn{bb: bytes.NewBuffer([]byte{})}

		conn, _ := apns.NewConn(apns.ProductionGateway, DummyCert, DummyKey)
		conn.NetConn = rwc

		It("should write out 'world!'", func() {
			conn.Write([]byte("world!"))
			Expect(rwc.bb.String()).To(Equal("world!"))
		})
	})

	Describe("#Close", func() {
		Context("with connection", func() {
			Context("no error", func() {
				rwc := mockTLSNetConn{bb: bytes.NewBuffer([]byte{})}

				conn, _ := apns.NewConn(apns.ProductionGateway, DummyCert, DummyKey)
				conn.NetConn = rwc

				It("should return no error", func() {
					Expect(rwc.Close()).To(BeNil())
				})
			})

			Context("with error", func() {
				rwc := mockTLSNetConn{bb: bytes.NewBuffer([]byte{})}

				conn, _ := apns.NewConn(apns.ProductionGateway, DummyCert, DummyKey)
				conn.NetConn = rwc

				rwc.err = io.EOF
				It("should return that error", func() {
					Expect(rwc.Close()).To(Equal(io.EOF))
				})
			})
		})

		Context("without connection", func() {
			c, _ := apns.NewConn(apns.ProductionGateway, DummyCert, DummyKey)
			It("should not return an error", func() {
				Expect(c.Close()).To(BeNil())
			})
		})
	})
})
