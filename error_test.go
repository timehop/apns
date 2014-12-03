package apns_test

import (
	"bytes"
	"encoding/binary"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/apns"
)

var _ = Describe("Error", func() {
	Describe(".NewError", func() {
		ShouldBeErrorWithErrStr := func(status int, errStr string) {
			var errPayload = func(command int, status int, identifier int) []byte {
				buffer := bytes.NewBuffer([]byte{})
				binary.Write(buffer, binary.BigEndian, uint8(command))
				binary.Write(buffer, binary.BigEndian, uint8(status))
				binary.Write(buffer, binary.BigEndian, uint32(identifier))
				return buffer.Bytes()
			}

			command := rand.Int()
			identifier := rand.Int()

			p := errPayload(command, status, identifier)
			e := apns.NewError(p)

			It("should parse the field out correctly", func() {
				Expect(e.Command).To(Equal(uint8(command)))
				Expect(e.Status).To(Equal(uint8(status)))
				Expect(e.Identifier).To(Equal(uint32(identifier)))
			})

			It("should have picked the right error string", func() {
				Expect(e.ErrStr).To(Equal(errStr))
			})
		}

		Context("processing error", func() {
			ShouldBeErrorWithErrStr(1, apns.ErrProcessing)
		})

		Context("device token error", func() {
			ShouldBeErrorWithErrStr(2, apns.ErrMissingDeviceToken)
		})

		Context("missing topic error", func() {
			ShouldBeErrorWithErrStr(3, apns.ErrMissingTopic)
		})

		Context("missing payload error", func() {
			ShouldBeErrorWithErrStr(4, apns.ErrMissingPayload)
		})

		Context("invalid token size error", func() {
			ShouldBeErrorWithErrStr(5, apns.ErrInvalidTokenSize)
		})

		Context("invalid topic size error", func() {
			ShouldBeErrorWithErrStr(6, apns.ErrInvalidTopicSize)
		})

		Context("invalid payload size error", func() {
			ShouldBeErrorWithErrStr(7, apns.ErrInvalidPayloadSize)
		})

		Context("invalid token error", func() {
			ShouldBeErrorWithErrStr(8, apns.ErrInvalidToken)
		})

		Context("shutdown error", func() {
			ShouldBeErrorWithErrStr(10, apns.ErrShutdown)
		})

		Context("unknown error", func() {
			ShouldBeErrorWithErrStr(255, apns.ErrUnknown)
		})

		Context("error with unrecognized code", func() {
			ShouldBeErrorWithErrStr(300, apns.ErrUnknown)
		})

		Context("not enough bytes", func() {
			It("should be ErrUnknown", func() {
				e := apns.NewError([]byte{})
				Expect(e).NotTo(BeNil())
				Expect(e.ErrStr).To(Equal(apns.ErrUnknown))
			})
		})
	})

	Describe("#Error", func() {
		It("should have an error string", func() {
			e := apns.Error{ErrStr: "this is an error string"}
			Expect(e.Error()).To(Equal("this is an error string"))
		})
	})
})
