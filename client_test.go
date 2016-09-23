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
	sendCB            func(n Notification) error
	requeueNotifs     []Notification
	disconnectedState bool
}

func (m *mockSession) Send(n Notification) error {
	if m.sendCB == nil {
		return nil
	}

	return m.sendCB(n)
}

func (m *mockSession) Connect() error {
	return nil
}

func (m *mockSession) RequeueableNotifications() []Notification {
	if len(m.requeueNotifs) == 0 {
		return []Notification{}
	}

	return m.requeueNotifs
}

func (m *mockSession) Disconnect() {
	m.disconnectedState = true
}

func (m *mockSession) Disconnected() bool {
	return m.disconnectedState
}

type badConnMockSession struct {
	*mockSession
}

func (m badConnMockSession) Connect() error {
	return errors.New("whatev")
}

var _ = Describe("Client", func() {
	BeforeEach(func() {
		newSession = func(_ Conn) Session { return &mockSession{} }
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
					newSession = func(_ Conn) Session {
						return &mockSession{
							sendCB: func(_ Notification) error {
								return errors.New("")
							},
						}
					}

					c, err := NewClient(SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
					Expect(err).To(BeNil())

					err = c.Send(Notification{DeviceToken: "lol"})
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Context("disconnected", func() {
			It("should reconnect", func() {
				newSessCount := 0
				newSession = func(_ Conn) Session {
					newSessCount++
					return &mockSession{}
				}

				c, err := NewClient(SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
				Expect(err).To(BeNil())

				c.sess.Disconnect()

				err = c.Send(Notification{DeviceToken: "0000000000000000000000000000000000000000000000000000000000000000"})
				Expect(err).To(BeNil())

				Expect(newSessCount).To(Equal(2))
			})
		})

		It("should reconnect and requeue", func() {
			newSessCount := 0
			sendCount := 0

			newSession = func(_ Conn) Session {
				newSessCount++
				return &mockSession{
					requeueNotifs: []Notification{
						Notification{},
						Notification{},
						Notification{},
					},
					sendCB: func(_ Notification) error {
						sendCount++
						return nil
					},
				}
			}

			c, err := NewClient(SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
			Expect(err).To(BeNil())

			c.sess.Disconnect()

			err = c.Send(Notification{DeviceToken: "0000000000000000000000000000000000000000000000000000000000000000"})
			Expect(err).To(BeNil())

			Expect(newSessCount).To(Equal(2))
			Expect(sendCount).To(Equal(4))
		})
	})
})
