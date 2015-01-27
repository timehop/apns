package apns_test

import (
	"io/ioutil"
	"net"
	"os"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/apns"
	"github.com/timehop/tcptest"
)

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
				_, err := apns.NewConn(apns.SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
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
				certFile.Write([]byte(tcptest.LocalhostCert))
				certFile.Close()

				keyFile, _ = ioutil.TempFile("", "key.pem")
				keyFile.Write([]byte(tcptest.LocalhostKey))
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
			Context("with untrusted certs", func() {
				It("should return an error", func(d Done) {
					s := tcptest.NewTLSServer(func(c net.Conn) {})
					defer s.Close()

					conn, err := apns.NewConn(s.Addr, "not trusted", "not even a little")
					Expect(err).NotTo(BeNil())

					err = conn.Connect()
					Expect(err).NotTo(BeNil())

					close(d)
				})
			})

			Context("trusting the certs", func() {
				It("should not return an error", func(d Done) {
					s := tcptest.NewUnstartedServer(func(c net.Conn) {
						defer c.Close()
						c.Write([]byte{}) // Connect
					})

					s.StartTLS()
					defer s.Close()

					conn, err := apns.NewConn(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
					Expect(err).To(BeNil())

					err = conn.Connect()
					Expect(err).To(BeNil())

					close(d)
				})
			})

			Context("with existing connection", func() {
				It("should not return an error", func(d Done) {
					s := tcptest.NewTLSServer(func(c net.Conn) {
						defer c.Close()
						c.Write([]byte{}) // Connect
					})
					defer s.Close()

					conn, _ := apns.NewConn(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))

					conn.Connect()

					err := conn.Connect()
					Expect(err).To(BeNil())

					close(d)
				})
			})
		})
	})

	Describe("#Read", func() {
		s := tcptest.NewTLSServer(func(c net.Conn) {
			defer c.Close()
			c.Write([]byte("hello!"))
		})
		defer s.Close()

		conn, _ := apns.NewConn(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
		conn.Connect()

		It("should read out 'hello!'", func() {
			p := make([]byte, 6)
			conn.Read(p)

			Expect(p).To(Equal([]byte("hello!")))
		})
	})

	Describe("#Write", func() {
		It("should read out 'hello!'", func(d Done) {
			s := tcptest.NewTLSServer(func(c net.Conn) {
				defer c.Close()
				c.Write([]byte{}) // Connect

				b := make([]byte, 6)
				c.Read(b)

				Expect(string(b)).To(Equal("hello!"))
				close(d)
			})
			defer s.Close()

			conn, _ := apns.NewConn(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
			conn.Connect()

			conn.Write([]byte("hello!"))
		})
	})

	Describe("#Close", func() {
		Context("with connection", func() {
			Context("no error", func() {
				It("should return no error", func() {
					s := tcptest.NewTLSServer(func(c net.Conn) {
						defer c.Close()
						c.Write([]byte{}) // Connect
					})
					defer s.Close()

					conn, _ := apns.NewConn(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
					conn.Connect()
					Expect(conn.Close()).To(BeNil())
				})
			})
		})

		Context("without connection", func() {
			It("should not return an error", func() {
				conn, _ := apns.NewConn("localhost:12345", string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
				Expect(conn.Close()).To(BeNil())
			})
		})
	})
})
