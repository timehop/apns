package apns

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type mockConn struct{}

func (m mockConn) Read(b []byte) (int, error) {
	return 0, nil
}

func (m mockConn) Write(b []byte) (int, error) {
	return 0, nil
}

func (m mockConn) Close() error {
	return nil
}

func (m mockConn) Connect() error {
	return nil
}

func (m mockConn) SetReadDeadline(deadline time.Time) error {
	return nil
}

var _ = Describe("Session", func() {
	Describe("NewSession", func() {
		It("creates a session", func() {
			s := NewSession(mockConn{})
			Expect(s).NotTo(BeNil())
		})
	})

	Describe("Connect", func() {
		Context("new state", func() {
			It("should not return an error", func() {
				s := NewSession(mockConn{})

				err := s.Connect()
				Expect(err).To(BeNil())
			})
		})

		Context("not new state", func() {
			It("should return an error", func() {
				sess := NewSession(mockConn{})

				s := sess.(*session)
				s.transitionState(sessionStateDisconnected)

				err := s.Connect()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Disconnected", func() {
		Context("not connected", func() {
			It("should not be true", func() {
				sess := NewSession(mockConn{})

				s := sess.(*session)
				s.transitionState(sessionStateDisconnected)

				Expect(s.Disconnected()).To(BeTrue())
			})
		})

		Context("connected", func() {
			It("should be false", func() {
				sess := NewSession(mockConn{})

				s := sess.(*session)
				s.Connect()

				Expect(s.Disconnected()).To(BeFalse())
			})
		})
	})

	Describe("Send", func() {
	})

	Describe("Disconnect", func() {
	})

	Describe("RequeueableNotifications", func() {
	})
})
