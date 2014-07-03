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

var _ = Describe("Feedback", func() {
	Describe(".NewFeedback", func() {
		Context("bad cert/key pair", func() {
			It("should error out", func() {
				_, err := apns.NewFeedback(apns.ProductionGateway, "missing", "missing_also")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("valid cert/key pair", func() {
			It("should create a valid client", func() {
				_, err := apns.NewFeedback(apns.ProductionGateway, DummyCert, DummyKey)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe(".NewFeedbackWithFiles", func() {
		Context("missing cert/key pair", func() {
			It("should error out", func() {
				_, err := apns.NewFeedbackWithFiles(apns.ProductionGateway, "missing", "missing_also")
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
				_, err := apns.NewFeedbackWithFiles(apns.ProductionGateway, certFile.Name(), keyFile.Name())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("#Receive", func() {
		Context("could not connect", func() {
			It("should not receive anything", func() {
				s := &mockTLSServer{}

				f, _ := apns.NewFeedback(s.Address(), DummyCert, DummyKey)
				f.Conn.Conf.InsecureSkipVerify = true

				c := f.Receive()

				r := 0
				for _ = range c {
					r += 1
				}

				Expect(r).To(Equal(0))
			})
		})

		Context("times out", func() {
			as := [][]serverAction{
				[]serverAction{
					serverAction{action: readAction, data: []byte{}},
				},
			}

			withMockServer(as, func(s *mockTLSServer) {
				f, _ := apns.NewFeedback(s.Address(), DummyCert, DummyKey)
				f.Conn.Conf.InsecureSkipVerify = true

				It("should not receive anything", func() {
					c := f.Receive()

					r := 0
					for _ = range c {
						r += 1
					}

					Expect(r).To(Equal(0))
				})
			})
		})

		Context("with feedback", func() {
			f1 := bytes.NewBuffer([]byte{})
			f2 := bytes.NewBuffer([]byte{})
			f3 := bytes.NewBuffer([]byte{})

			binary.Write(f1, binary.BigEndian, uint32(1404358249))
			binary.Write(f1, binary.BigEndian, uint16(32))
			binary.Write(f1, binary.BigEndian, []byte("f111a111121112111111111411111111"))

			binary.Write(f2, binary.BigEndian, uint32(1404352249))
			binary.Write(f2, binary.BigEndian, uint16(32))
			binary.Write(f2, binary.BigEndian, []byte("f111a11112111bbbbb11111411111111"))

			binary.Write(f3, binary.BigEndian, uint32(1394352249))
			binary.Write(f3, binary.BigEndian, uint16(32))
			binary.Write(f3, binary.BigEndian, []byte("f111a11112111bbbbb11111499999999"))

			as := [][]serverAction{
				[]serverAction{
					serverAction{action: writeAction, data: f1.Bytes()},
					serverAction{action: writeAction, data: f2.Bytes()},
					serverAction{action: writeAction, data: f3.Bytes()},
				},
			}

			It("should receive feedback", func(d Done) {
				withMockServer(as, func(s *mockTLSServer) {
					f, _ := apns.NewFeedback(s.Address(), DummyCert, DummyKey)
					f.Conn.Conf.InsecureSkipVerify = true

					c := f.Receive()

					r1 := <-c
					Expect(r1.Timestamp).To(Equal(time.Unix(1404358249, 0)))
					Expect(r1.TokenLength).To(Equal(uint16(32)))
					Expect(r1.DeviceToken).To(Equal("f111a111121112111111111411111111"))

					r2 := <-c
					Expect(r2.Timestamp).To(Equal(time.Unix(1404352249, 0)))
					Expect(r2.TokenLength).To(Equal(uint16(32)))
					Expect(r2.DeviceToken).To(Equal("f111a11112111bbbbb11111411111111"))

					r3 := <-c
					Expect(r3.Timestamp).To(Equal(time.Unix(1394352249, 0)))
					Expect(r3.TokenLength).To(Equal(uint16(32)))
					Expect(r3.DeviceToken).To(Equal("f111a11112111bbbbb11111499999999"))

					<-c
					close(d)
				})
			})
		})
	})
})
