package apns_test

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/apns"
)

var _ = Describe("Client", func() {
	Describe(".NewConn", func() {
		Context("bad cert/key pair", func() {
			It("should error out", func() {
				_, err := apns.NewClient(apns.ProductionGateway, "missing", "missing_also")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("valid cert/key pair", func() {
			It("should create a valid client", func() {
				c, err := apns.NewClient(apns.ProductionGateway, DummyCert, DummyKey)
				Expect(err).To(BeNil())
				Expect(c.Conn).NotTo(BeNil())
			})
		})
	})

	Describe(".NewConnWithFiles", func() {
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

			It("should create a valid client", func() {
				c, err := apns.NewClientWithFiles(apns.ProductionGateway, certFile.Name(), keyFile.Name())
				Expect(err).To(BeNil())
				Expect(c.Conn).NotTo(BeNil())
			})
		})
	})

	Describe("#Send", func() {
		Context("simple write", func() {
			as := [][]serverAction{
				[]serverAction{
					serverAction{action: readAction, data: []byte{}},
				},
			}

			It("should not return an error", func(d Done) {
				mockDone := make(chan interface{})
				withMockServerAsync(as, mockDone, func(s *mockTLSServer) {
					c, _ := apns.NewClient(s.Address(), DummyCert, DummyKey)
					c.Conn.Conf.InsecureSkipVerify = true

					Expect(c.Send(apns.Notification{})).To(BeNil())

					close(mockDone)
					close(d)
				})
			})
		})

		Context("simple write with buffer", func() {
			as := [][]serverAction{
				[]serverAction{
					serverAction{action: readAction, data: []byte{}},
				},
			}

			It("should not return an error", func(d Done) {
				mockDone := make(chan interface{})
				withMockServerAsync(as, mockDone, func(s *mockTLSServer) {
					c, _ := apns.NewClient(s.Address(), DummyCert, DummyKey)
					c.Conn.Conf.InsecureSkipVerify = true

					for i := 0; i < 54; i++ {
						Expect(c.Send(apns.Notification{})).To(BeNil())
					}

					close(mockDone)
					close(d)
				})
			})
		})

		Context("multiple write", func() {
			as := [][]serverAction{
				[]serverAction{
					serverAction{action: readAction, data: []byte{}},
					serverAction{action: readAction, data: []byte{}},
				},
			}

			It("should not return an error", func(d Done) {
				mockDone := make(chan interface{})
				withMockServerAsync(as, mockDone, func(s *mockTLSServer) {
					c, _ := apns.NewClient(s.Address(), DummyCert, DummyKey)
					c.Conn.Conf.InsecureSkipVerify = true

					Expect(c.Send(apns.Notification{})).To(BeNil())
					Expect(c.Send(apns.Notification{})).To(BeNil())

					close(mockDone)
					close(d)
				})
			})
		})

		Context("bad push", func() {
			n := apns.Notification{Identifier: 9, ID: "some_rando"}
			nb, _ := n.ToBinary()
			nbcb := make([]byte, len(nb))

			errPayload := bytes.NewBuffer([]byte{})
			binary.Write(errPayload, binary.BigEndian, uint8(8))
			binary.Write(errPayload, binary.BigEndian, uint8(8))
			binary.Write(errPayload, binary.BigEndian, uint32(9))

			as := [][]serverAction{
				[]serverAction{
					serverAction{action: readAction, data: []byte{}},
					serverAction{action: readAction, data: nbcb, cb: func(a serverAction) {
						Expect(a.data).To(Equal(nb))
					}},

					// Bad push results in a close
					serverAction{action: writeAction, data: errPayload.Bytes()},
					serverAction{action: closeAction, data: []byte{}},
				},
			}

			It("should not return an error", func(d Done) {
				mockDone := make(chan interface{})
				withMockServerAsync(as, mockDone, func(s *mockTLSServer) {
					c, _ := apns.NewClient(s.Address(), DummyCert, DummyKey)
					c.Conn.Conf.InsecureSkipVerify = true

					go func() {
						n := <-c.FailedNotifs

						Expect(n.Notif.Identifier).To(Equal(uint32(9)))
						Expect(n.Notif.ID).To(Equal("some_rando"))

						close(mockDone)
						close(d)
					}()

					Expect(c.Send(n)).To(BeNil())
				})
			})
		})

		Context("closed, reconnect", func() {
			done := make(chan bool)

			n1 := apns.Notification{Identifier: 1}
			n1b, _ := n1.ToBinary()
			n1bcb := make([]byte, len(n1b))

			errPayload := bytes.NewBuffer([]byte{})
			binary.Write(errPayload, binary.BigEndian, uint8(8))
			binary.Write(errPayload, binary.BigEndian, uint8(8))
			binary.Write(errPayload, binary.BigEndian, uint32(2))

			It("should not return an error", func(d Done) {
				mockDone := make(chan interface{})

				as := [][]serverAction{
					[]serverAction{
						// Write error
						serverAction{action: writeAction, data: errPayload.Bytes(), cb: func(a serverAction) {
							done <- true
						}},

						// Close on error
						serverAction{action: closeAction, cb: func(a serverAction) {
						}},
					},
					[]serverAction{
						// Reconnect
						serverAction{action: readAction, data: []byte{}, cb: func(a serverAction) {
							// Reconnected
						}},

						// Read first good notification
						serverAction{action: readAction, data: n1bcb, cb: func(a serverAction) {
							Expect(a.data).To(Equal(n1b))

							close(mockDone)
							close(d)
						}},
					},
				}

				withMockServerAsync(as, mockDone, func(s *mockTLSServer) {
					c, _ := apns.NewClient(s.Address(), DummyCert, DummyKey)
					c.Conn.Conf.InsecureSkipVerify = true

					<-done
					time.Sleep(5 * time.Millisecond)

					// Good
					Expect(c.Send(n1)).To(BeNil())
				})
			})
		})

		Context("good, close, good, requeue of last good", func() {
			closed := make(chan bool)

			n1 := apns.Notification{Identifier: 1}
			n2 := apns.Notification{Identifier: 2}

			n1b, _ := n1.ToBinary()
			n2b, _ := n2.ToBinary()

			n1bcb := make([]byte, len(n1b))
			n2bcb := make([]byte, len(n2b))

			It("should not return an error", func(d Done) {
				mockDone := make(chan interface{})
				as := [][]serverAction{
					[]serverAction{
						// Connect
						serverAction{action: readAction, data: []byte{}, cb: func(a serverAction) {
							// Handshake
						}},

						// Read first good notification
						serverAction{action: readAction, data: n1bcb, cb: func(a serverAction) {
							Expect(a.data).To(Equal(n1b))
						}},

						// Close on error
						serverAction{action: closeAction, cb: func(a serverAction) {
							closed <- true
						}},
					},
					[]serverAction{
						// Reconnect
						serverAction{action: readAction, data: []byte{}, cb: func(a serverAction) {
							// Reconnected
						}},

						// Requeue
						serverAction{action: readAction, data: n2bcb, cb: func(a serverAction) {
							Expect(a.data).To(Equal(n2b))

							close(mockDone)
							close(d)
						}},
					},
				}

				withMockServerAsync(as, mockDone, func(s *mockTLSServer) {
					c, _ := apns.NewClient(s.Address(), DummyCert, DummyKey)
					c.Conn.Conf.InsecureSkipVerify = true

					// Good
					Expect(c.Send(n1)).To(BeNil())

					<-closed
					time.Sleep(5 * time.Millisecond)

					// Good
					Expect(c.Send(n2)).To(BeNil())
				})
			})
		})

		Context("good, bad, good, requeue of last good", func() {
			It("should not return an error", func(d Done) {
				mockDone := make(chan interface{})

				n1 := apns.Notification{Identifier: 1}
				n2 := apns.Notification{Identifier: 2}
				n3 := apns.Notification{Identifier: 3}

				n1b, _ := n1.ToBinary()
				n2b, _ := n2.ToBinary()
				n3b, _ := n3.ToBinary()

				n1bcb := make([]byte, len(n1b))
				n2bcb := make([]byte, len(n2b))
				n3bcb := make([]byte, len(n3b))

				errPayload := bytes.NewBuffer([]byte{})
				binary.Write(errPayload, binary.BigEndian, uint8(8))
				binary.Write(errPayload, binary.BigEndian, uint8(8))
				binary.Write(errPayload, binary.BigEndian, uint32(2))

				as := [][]serverAction{
					[]serverAction{
						// Connect
						serverAction{action: readAction, data: []byte{}, cb: func(a serverAction) {
							// Handshake
						}},

						// Read first good notification
						serverAction{action: readAction, data: n1bcb, cb: func(a serverAction) {
							Expect(a.data).To(Equal(n1b))
						}},

						// Read bad notification
						serverAction{action: readAction, data: n2bcb, cb: func(a serverAction) {
							Expect(a.data).To(Equal(n2b))
						}},

						// Read second good notification
						serverAction{action: readAction, data: n3bcb, cb: func(a serverAction) {
							Expect(a.data).To(Equal(n3b))
						}},

						// Write error
						serverAction{action: writeAction, data: errPayload.Bytes(), cb: func(a serverAction) {
						}},

						// Close on error
						serverAction{action: closeAction, cb: func(a serverAction) {
						}},
					},
					[]serverAction{
						// Reconnect
						serverAction{action: readAction, data: []byte{}, cb: func(a serverAction) {
							// Reconnected
						}},

						// Requeue
						serverAction{action: readAction, data: n3bcb, cb: func(a serverAction) {
							Expect(a.data).To(Equal(n3b))

							close(mockDone)
							close(d)
						}},
					},
				}

				withMockServerAsync(as, mockDone, func(s *mockTLSServer) {
					c, _ := apns.NewClient(s.Address(), DummyCert, DummyKey)
					c.Conn.Conf.InsecureSkipVerify = true

					// Good
					Expect(c.Send(n1)).To(BeNil())

					// Bad
					Expect(c.Send(n2)).To(BeNil())

					// Good
					Expect(c.Send(n3)).To(BeNil())
				})
			})
		})
	})
})
