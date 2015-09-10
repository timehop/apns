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
			Expect(b.Number).To(Equal(uint(0)))
			Expect(b.IsSet).To(BeFalse())
		})
	})

	Describe("NewBadgeNumber", func() {
		var b apns.BadgeNumber

		Context("with an argument", func() {
			It("should have values set properly", func() {
				b.Set(5)
				Expect(b.IsSet).To(BeTrue())
				Expect(b.Number).To(Equal(uint(5)))
			})
		})
		Context("when unset", func() {
			It("should reset its values", func() {
				b.Unset()
				Expect(b.IsSet).To(BeFalse())
				Expect(b.Number).To(Equal(uint(0)))
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
				bn := apns.BadgeNumber{}
				bn.Set(10)
				_, err := json.Marshal(BadgeNumbers{
					A: bn,
				})
				Expect(err).To(BeNil())
			})
			It("should create the proper values", func() {
				bn := apns.BadgeNumber{}
				bn.Set(10)
				bns := BadgeNumbers{
					A: bn,
				}
				b, _ := json.Marshal(bns)
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
				Expect(bnumbers.A.IsSet).To(BeTrue())
				Expect(bnumbers.B.IsSet).To(BeTrue())

				Expect(bnumbers.A.Number).To(Equal(uint(10)))
				Expect(bnumbers.B.Number).To(Equal(uint(0)))
			})
		})
	})
})
