package tcp

import (
	"net"
	"sync"
	"testing"
)

func TestNewClient(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newClient(clientConn, 100)

	if c.maxBytes != 100 {
		t.Errorf("expected bandwidth limit 100, got %d", c.maxBytes)
	}

	if c.sentBytes != 0 {
		t.Errorf("expected sentbytes 0, got %d", c.sentBytes)
	}

	if c.sendCh == nil {
		t.Error("sendCh should not be nil")
	}

	if cap(c.sendCh) != 10 {
		t.Errorf("sendCh should contain exactly 10 bytes, got %d", cap(c.sendCh))
	}
}

func TestCanLimitBandwidth(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newClient(clientConn, 100)

	c.incrementSentBytes(10)
	if c.isBandwidthLimitExceeded() {
		t.Errorf("expected bandwidth limit not to be exceeded (limit: %d, used: %d), but it was", c.maxBytes, c.sentBytes)
	}

	c.incrementSentBytes(89)
	if c.isBandwidthLimitExceeded() {
		t.Errorf("expected bandwidth limit not to be exceeded after adding 99 total bytes (limit: %d, used: %d)", c.maxBytes, c.sentBytes)
	}

	c.incrementSentBytes(1)
	if !c.isBandwidthLimitExceeded() {
		t.Errorf("expected bandwidth limit to be exceeded after adding 100 total bytes (limit: %d, used: %d)", c.maxBytes, c.sentBytes)
	}
}

func TestConcurrentIncrement(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newClient(clientConn, 1000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.incrementSentBytes(10)
		}()
	}
	wg.Wait()
	if c.sentBytes != 1000 {
		t.Errorf("expected sentBytes to be 1000, got %d", c.sentBytes)
	}
}
