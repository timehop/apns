package apns_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timehop/apns"
)

var _ = Describe("BadgeNumber", func() {
	Describe("Defaults", func() {
		It("Should have the proper default values", func() {
			b := apns.BadgeNumber{}
			Expect(b.Number).To(Equal(0))
			Expect(b.Valid).To(BeFalse())
		})
	})

	Describe("NewBadgeNumber", func() {
		var b apns.BadgeNumber

		Context("with an argument", func() {
			It("should have values set properly", func() {
				b = apns.NewBadgeNumber(5)
				Expect(b.Valid).To(BeTrue())
				Expect(b.Number).To(Equal(5))
			})
		})
		Context("when unset", func() {
			It("should reset its values", func() {
				b.Unset()
				Expect(b.Valid).To(BeFalse())
				Expect(b.Number).To(Equal(0))
			})
		})
	})

	Describe("JSON handling", func() {
		type BadgeNumbers struct {
			A apns.BadgeNumber `json:"a"`
			B apns.BadgeNumber `json:"b"`
		}

		Context("when marshalling", func() {
			It("should not error", func() {
				_, err := json.Marshal(BadgeNumbers{
					A: apns.NewBadgeNumber(10),
				})
				Expect(err).To(BeNil())
			})
			It("should create the proper values", func() {
				bn := BadgeNumbers{
					A: apns.NewBadgeNumber(10),
				}
				b, _ := json.Marshal(bn)
				expected := "{\"a\":10,\"b\":0}"
				Expect(string(b)).To(Equal(expected))
			})
		})

		Context("when unmarshalled", func() {
			var bnumbers BadgeNumbers

			It("should unmarshal without errors", func() {
				err := json.Unmarshal([]byte("{\"a\":10,\"b\":0}"), &bnumbers)
				Expect(err).To(BeNil())
			})
			It("should populate the struct properly", func() {
				json.Unmarshal([]byte("{\"a\":10,\"b\":0}"), &bnumbers)
				Expect(bnumbers.A.Valid).To(BeTrue())
				Expect(bnumbers.B.Valid).To(BeTrue())

				Expect(bnumbers.A.Number).To(Equal(10))
				Expect(bnumbers.B.Number).To(Equal(0))
			})
		})
	})
})
