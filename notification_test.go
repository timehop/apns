package apns_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/apns"
)

var _ = Describe("Notifications", func() {
	Describe("Alert", func() {
		Describe("JSON marshalling", func() {
			Context("only body", func() {
				It("should just have that field", func() {
					a := apns.Alert{Body: "whatup"}

					j, err := json.Marshal(a)

					Expect(err).To(BeNil())
					Expect(j).To(Equal([]byte(`{"body":"whatup"}`)))
				})
			})

			Context("only loc-key", func() {
				It("should just have that field", func() {
					a := apns.Alert{LocKey: "localization"}

					j, err := json.Marshal(a)

					Expect(err).To(BeNil())
					Expect(j).To(Equal([]byte(`{"loc-key":"localization"}`)))
				})
			})

			Context("only loc-args", func() {
				It("should just have that field", func() {
					a := apns.Alert{LocArgs: []string{"world", "cup"}}

					j, err := json.Marshal(a)

					Expect(err).To(BeNil())
					Expect(j).To(Equal([]byte(`{"loc-args":["world","cup"]}`)))
				})
			})

			Context("only action-loc-key", func() {
				It("should just have that field", func() {
					a := apns.Alert{ActionLocKey: "akshun localization"}

					j, err := json.Marshal(a)

					Expect(err).To(BeNil())
					Expect(j).To(Equal([]byte(`{"action-loc-key":"akshun localization"}`)))
				})
			})

			Context("only launch image", func() {
				It("should just have that field", func() {
					a := apns.Alert{LaunchImage: "dee fault"}

					j, err := json.Marshal(a)

					Expect(err).To(BeNil())
					Expect(j).To(Equal([]byte(`{"launch-image":"dee fault"}`)))
				})
			})

			Context("fully loaded", func() {
				It("should serialize", func() {
					a := apns.Alert{Body: "USA scores!", LocKey: "game", LocArgs: []string{"USA", "BRA"}, LaunchImage: "scoreboard"}

					j, err := json.Marshal(a)

					Expect(err).To(BeNil())
					Expect(j).To(Equal([]byte(`{"body":"USA scores!","loc-key":"game","loc-args":["USA","BRA"],"launch-image":"scoreboard"}`)))
				})
			})
		})
	})

	Describe("Safari", func() {
		Describe("#MarshalJSON", func() {
			Context("with complete payload", func() {
				It("should marshal APS", func() {
					p := apns.NewPayload()

					p.APS.Alert.Title = "Hello World!"
					p.APS.Alert.Body = "This is a body"
					p.APS.Alert.Action = "Launch"
					p.APS.URLArgs = []string{"hello", "world"}

					b, err := json.Marshal(p)

					Expect(err).To(BeNil())
					Expect(b).To(Equal([]byte(`{"aps":{"alert":{"body":"This is a body","title":"Hello World!","action":"Launch"},"url-args":["hello","world"]}}`)))
				})
			})
		})
	})

	Describe("Payload", func() {
		Describe("#MarshalJSON", func() {
			Context("no alert (as with Passbook)", func() {
				It("should not contain the alert struct", func() {
					p := apns.NewPayload()
					b, err := json.Marshal(p)

					Expect(err).To(BeNil())
					Expect(b).To(Equal([]byte(`{"aps":{}}`)))
				})
			})
			Context("no alert with content available (as with Newsstand)", func() {
				It("should not contain the alert struct", func() {
					p := apns.NewPayload()
					p.APS.ContentAvailable = 1
					b, err := json.Marshal(p)

					Expect(err).To(BeNil())
					Expect(b).To(Equal([]byte(`{"aps":{"content-available":1}}`)))
				})
			})
			Context("with only APS", func() {
				It("should marshal APS", func() {
					p := apns.NewPayload()

					p.APS.Alert.Body = "testing"

					b, err := json.Marshal(p)

					Expect(err).To(BeNil())
					Expect(b).To(Equal([]byte(`{"aps":{"alert":"testing"}}`)))
				})
			})

			Context("with custom attributes APS", func() {
				It("should marshal APS", func() {
					p := apns.NewPayload()

					p.APS.Alert.Body = "testing"
					p.SetCustomValue("email", "come@me.bro")

					b, err := json.Marshal(p)

					Expect(err).To(BeNil())
					Expect(b).To(Equal([]byte(`{"aps":{"alert":"testing"},"email":"come@me.bro"}`)))
				})
			})

			Context("with only MDM", func() {
				It("should marshal MDM", func() {
					p := apns.NewPayload()

					p.MDM = "00000000-1111-3333-4444-555555555555"

					b, err := json.Marshal(p)

					Expect(err).To(BeNil())
					Expect(b).To(Equal([]byte(`{"mdm":"00000000-1111-3333-4444-555555555555"}`)))
				})
			})
		})
	})

	Describe("APS", func() {
		Context("badge with a zero (clears notifications)", func() {
			It("should contain zero", func() {
				zero := 0
				a := apns.APS{Badge: &zero}

				j, err := json.Marshal(a)

				Expect(err).To(BeNil())
				Expect(j).To(Equal([]byte(`{"badge":0}`)))
			})
		})
		Context("no badge specified (do nothing)", func() {
			It("should omit the badge field", func() {
				a := apns.APS{}

				j, err := json.Marshal(a)

				Expect(err).To(BeNil())
				Expect(j).To(Equal([]byte(`{}`)))
			})
		})
	})

	Describe("Notification", func() {
		Describe("#ToBinary", func() {
			Context("invalid token format", func() {
				n := apns.NewNotification()
				n.DeviceToken = "totally not a valid token"

				It("should return an error", func() {
					_, err := n.ToBinary()
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(ContainSubstring("convert token to hex error"))
				})
			})

			Context("valid payload", func() {
				It("should generate the correct byte payload with expiry", func() {
					t := time.Unix(1404102833, 0)

					n := apns.NewNotification()
					n.Identifier = uint32(123123)
					n.DeviceToken = "9999999999999999999999999999999999999999999999999999999999999999"
					n.Priority = apns.PriorityImmediate
					n.Expiration = &t

					b, err := n.ToBinary()

					Expect(err).To(BeNil())

					buf := bytes.NewBuffer(b)

					var expiry uint32

					buf.Next(1 + 4 + 1 + 2 + 32 + 1 + 2 + 20 + 1 + 2 + 4 + 1 + 2)

					// Expiry
					binary.Read(buf, binary.BigEndian, &expiry)
				})

				It("should generate the correct byte payload", func() {
					n := apns.NewNotification()
					n.Identifier = uint32(123123)
					n.DeviceToken = "9999999999999999999999999999999999999999999999999999999999999999"
					n.Priority = apns.PriorityImmediate
					b, err := n.ToBinary()

					Expect(err).To(BeNil())

					buf := bytes.NewBuffer(b)

					var command, tokID, payloadID, identifierID, expiryID, priorityID uint8
					var tokLen, payloadLen, identifierLen, expiryLen, priorityLen uint16
					var frameLen, identifier, expiry uint32
					var priority byte
					var tok [32]byte
					var payload [10]byte

					binary.Read(buf, binary.BigEndian, &command)
					binary.Read(buf, binary.BigEndian, &frameLen)

					// Token
					binary.Read(buf, binary.BigEndian, &tokID)
					binary.Read(buf, binary.BigEndian, &tokLen)
					binary.Read(buf, binary.BigEndian, &tok)

					// Payload
					binary.Read(buf, binary.BigEndian, &payloadID)
					binary.Read(buf, binary.BigEndian, &payloadLen)
					binary.Read(buf, binary.BigEndian, &payload)

					// Identifier
					binary.Read(buf, binary.BigEndian, &identifierID)
					binary.Read(buf, binary.BigEndian, &identifierLen)
					binary.Read(buf, binary.BigEndian, &identifier)

					// Expiry
					binary.Read(buf, binary.BigEndian, &expiryID)
					binary.Read(buf, binary.BigEndian, &expiryLen)
					binary.Read(buf, binary.BigEndian, &expiry)

					// Priority
					binary.Read(buf, binary.BigEndian, &priorityID)
					binary.Read(buf, binary.BigEndian, &priorityLen)
					binary.Read(buf, binary.BigEndian, &priority)

					Expect(command).To(Equal(uint8(2)))
					Expect(frameLen).To(Equal(uint32(66)))

					// Token
					Expect(tokID).To(Equal(uint8(1)))
					Expect(tokLen).To(Equal(uint16(32)))
					Expect(tok).To(Equal([32]byte{153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153}))

					// Payload
					Expect(payloadID).To(Equal(uint8(2)))
					Expect(payloadLen).To(Equal(uint16(10)))
					Expect(payload).To(Equal([10]byte{123, 34, 97, 112, 115, 34, 58, 123, 125, 125}))

					// Identifier
					Expect(identifierID).To(Equal(uint8(3)))
					Expect(identifierLen).To(Equal(uint16(4)))
					Expect(identifier).To(Equal(uint32(123123)))

					// Expiry
					Expect(expiryID).To(Equal(uint8(4)))
					Expect(expiryLen).To(Equal(uint16(4)))
					Expect(expiry).To(Equal(uint32(0)))

					// Priority
					Expect(priorityID).To(Equal(uint8(5)))
					Expect(priorityLen).To(Equal(uint16(1)))
					Expect(priority).To(Equal(uint8(10)))
				})
			})
		})
	})
})
