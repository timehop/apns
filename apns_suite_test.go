package apns_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"time"
)

type mockConn struct {
	connect         func() error
	read            func([]byte) (int, error)
	setReadDeadline func(time.Time) error
}

func (m *mockConn) Connect() error {
	if m.connect != nil {
		return m.connect()
	}

	return nil
}

func (m *mockConn) Read(b []byte) (int, error) {
	if m.read != nil {
		return m.read(b)
	}
	return 0, nil
}

func (m *mockConn) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	if m.setReadDeadline != nil {
		return m.setReadDeadline(t)
	}
	return nil
}

func TestApns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apns Suite")
}
