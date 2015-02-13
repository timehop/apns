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

var _ = Describe("Client", func() {
	Describe(".NewClient", func() {
		Context("bad cert/key pair", func() {
			It("should error out", func() {
				_, err := apns.NewClient(apns.ProductionGateway, "missing", "missing_also")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("valid cert/key pair", func() {
			It("should create a valid client", func() {
				_, err := apns.NewClient(apns.SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
				Expect(err).To(BeNil())
			})
		})
	})

	Describe(".NewClientWithFiles", func() {
		Context("missing cert/key pair", func() {
			It("should error out", func() {
				_, err := apns.NewClientWithFiles(apns.ProductionGateway, "missing", "missing_also")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("valid cert/key pair", func() {
			var certFile, keyFile *os.File

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

			It("should create a valid client", func() {
				_, err := apns.NewClientWithFiles(apns.ProductionGateway, certFile.Name(), keyFile.Name())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Connect", func() {
		It("should not return an error", func() {
			s := tcptest.NewTLSServer(func(c net.Conn) {
				c.Write([]byte{0})
				c.Close()
			})
			defer s.Close()

			c, err := apns.NewClient(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
			Expect(err).To(BeNil())

			err = c.Connect()
			Expect(err).To(BeNil())
		})
	})

	Describe("Send", func() {
		Context("valid push", func() {
			It("should not return an error", func() {
				s := tcptest.NewTLSServer(func(c net.Conn) {
					c.Write([]byte{0})
					c.Close()
				})
				defer s.Close()

				c, err := apns.NewClient(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
				Expect(err).To(BeNil())

				err = c.Connect()
				Expect(err).To(BeNil())

				err = c.Send(apns.Notification{DeviceToken: "0000000000000000000000000000000000000000000000000000000000000000"})
				Expect(err).To(BeNil())
			})
		})

		Context("invalid notification", func() {
			It("should not return an error", func() {
				s := tcptest.NewTLSServer(func(c net.Conn) {
					c.Write([]byte{0})
					c.Close()
				})
				defer s.Close()

				c, err := apns.NewClient(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
				Expect(err).To(BeNil())

				err = c.Connect()
				Expect(err).To(BeNil())

				err = c.Send(apns.Notification{DeviceToken: "lol"})
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
