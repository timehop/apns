package apns

import (
	"errors"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/tcptest"
)

type mockSession struct {
	sendErr error
}

func (m mockSession) Send(n Notification) error {
	return m.sendErr
}

func (m mockSession) Connect() error {
	return nil
}

func (m mockSession) RequeueableNotifications() []Notification {
	return []Notification{}
}

func (m mockSession) Disconnect() {
}

func (m mockSession) Disconnected() bool {
	return false
}

type badConnMockSession struct {
	mockSession
}

func (_ badConnMockSession) Connect() error {
	return errors.New("whatev")
}

var _ = Describe("Client", func() {
	BeforeEach(func() {
		newSession = func(_ Conn) Session { return mockSession{} }
	})

	Describe(".NewClient", func() {
		Context("bad cert/key pair", func() {
			It("should error out", func() {
				_, err := NewClient(ProductionGateway, "missing", "missing_also")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("valid cert/key pair", func() {
			It("should create a valid client", func() {
				_, err := NewClient(SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
				Expect(err).To(BeNil())
			})
		})

		Context("bad connection", func() {
			It("should error out", func() {
				newSession = func(_ Conn) Session { return badConnMockSession{} }

				_, err := NewClient(SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe(".NewClientWithFiles", func() {
		Context("missing cert/key pair", func() {
			It("should error out", func() {
				_, err := NewClientWithFiles(ProductionGateway, "missing", "missing_also")
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
				_, err := NewClientWithFiles(ProductionGateway, certFile.Name(), keyFile.Name())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Send", func() {
		Context("connected", func() {
			Context("valid push", func() {
				It("should not return an error", func() {
					c, err := NewClient(SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
					Expect(err).To(BeNil())

					err = c.Send(Notification{DeviceToken: "0000000000000000000000000000000000000000000000000000000000000000"})
					Expect(err).To(BeNil())
				})
			})

			Context("invalid notification", func() {
				It("should return an error", func() {
					newSession = func(_ Conn) Session { return mockSession{sendErr: errors.New("")} }

					c, err := NewClient(SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
					Expect(err).To(BeNil())

					err = c.Send(Notification{DeviceToken: "lol"})
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Context("disconnected", func() {
		})
	})
})
