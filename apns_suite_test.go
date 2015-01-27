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
	readWithTimeout func([]byte, time.Time) (int, error)
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

func (m *mockConn) ReadWithTimeout(b []byte, t time.Time) (int, error) {
	if m.readWithTimeout != nil {
		return m.readWithTimeout(b, t)
	}

	return 0, nil
}

func TestApns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apns Suite")
}
