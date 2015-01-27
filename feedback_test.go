package apns_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/apns"
	"github.com/timehop/tcptest"
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
				_, err := apns.NewFeedback(apns.SandboxGateway, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
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
				_, err := apns.NewFeedbackWithFiles(apns.ProductionGateway, certFile.Name(), keyFile.Name())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("#Receive", func() {
		Context("could not connect", func() {
			It("should not receive anything", func() {
				m := mockConn{
					connect: func() error {
						return io.EOF
					},
				}

				f := apns.Feedback{Conn: &m}
				c := f.Receive()

				r := 0
				for _ = range c {
					r += 1
				}

				Expect(r).To(Equal(0))
			})
		})

		Context("times out", func() {
			It("should not receive anything", func() {
				m := mockConn{
					readWithTimeout: func(b []byte, t time.Time) (int, error) {
						return 0, net.UnknownNetworkError("")
					},
				}

				f := apns.Feedback{Conn: &m}
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

		// The final token strings
		t1 := "00a18269661e9406aea59a5620b05c7c0e371574fa6f251951de8d7a5a292535"
		t2 := "00a1a4b7294fcfbc5293f63d4298fcecd9c20a893befd45adceead5fc92d3319"
		t3 := "00a1b7893d5e85eb8bb7bf0846b464d075248555118ae893b06e96cfb8d678e3"

		bt1, _ := hex.DecodeString(t1)
		bt2, _ := hex.DecodeString(t2)
		bt3, _ := hex.DecodeString(t3)

		binary.Write(f1, binary.BigEndian, uint32(1404358249))
		binary.Write(f1, binary.BigEndian, uint16(len(bt1)))
		binary.Write(f1, binary.BigEndian, bt1)

		binary.Write(f2, binary.BigEndian, uint32(1404352249))
		binary.Write(f2, binary.BigEndian, uint16(len(bt2)))
		binary.Write(f2, binary.BigEndian, bt2)

		binary.Write(f3, binary.BigEndian, uint32(1394352249))
		fmt.Println("f3 bytes", f3)

		binary.Write(f3, binary.BigEndian, uint16(len(bt3)))
		binary.Write(f3, binary.BigEndian, bt3)

		It("should receive feedback", func(d Done) {
			s := tcptest.NewTLSServer(func(c net.Conn) {
				c.Write(f1.Bytes())
				c.Write(f2.Bytes())
				c.Write(f3.Bytes())

				// TODO(bw) figure out why we need this
				c.Write([]byte{0})
				c.Close()
			})
			defer s.Close()

			f, err := apns.NewFeedback(s.Addr, string(tcptest.LocalhostCert), string(tcptest.LocalhostKey))
			Expect(err).To(BeNil())

			c := f.Receive()

			r1 := <-c
			Expect(r1.Timestamp.Unix()).To(Equal(int64(1404358249)))
			Expect(r1.TokenLength).To(Equal(uint16(len(bt1))))
			Expect(r1.DeviceToken).To(Equal(t1))

			r2 := <-c
			Expect(r2.Timestamp.Unix()).To(Equal(int64(1404352249)))
			Expect(r2.TokenLength).To(Equal(uint16(len(bt2))))
			Expect(r2.DeviceToken).To(Equal(t2))

			r3 := <-c
			Expect(r3.Timestamp.Unix()).To(Equal(int64(1394352249)))
			Expect(r3.TokenLength).To(Equal(uint16(len(bt3))))
			Expect(r3.DeviceToken).To(Equal(t3))

			<-c
			close(d)
		})
	})
})
