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

	if c.uploadedBytes != 0 {
		t.Errorf("expected uploadedBytes 0, got %d", c.uploadedBytes)
	}

	if c.downloadedBytes != 0 {
		t.Errorf("expected downloadedBytes 0, got %d", c.downloadedBytes)
	}

	if c.sendCh == nil {
		t.Error("sendCh should not be nil")
	}

	if cap(c.sendCh) != 10 {
		t.Errorf("sendCh should contain exactly 10 bytes, got %d", cap(c.sendCh))
	}
}

func TestCanLimitUploadedBytes(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newClient(clientConn, 100)

	c.incrementUploadedBytes(10)
	if c.isUploadedBytesLimitExceeded() {
		t.Errorf("expected upload limit not to be exceeded (limit: %d, used: %d), but it was", c.maxBytes, c.uploadedBytes)
	}

	c.incrementUploadedBytes(89)
	if c.isUploadedBytesLimitExceeded() {
		t.Errorf("expected upload limit not to be exceeded after adding 99 total bytes (limit: %d, used: %d)", c.maxBytes, c.uploadedBytes)
	}

	c.incrementUploadedBytes(1)
	if !c.isUploadedBytesLimitExceeded() {
		t.Errorf("expected upload limit to be exceeded after adding 100 total bytes (limit: %d, used: %d)", c.maxBytes, c.uploadedBytes)
	}
}

func TestCanLimitDownloadBytes(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newClient(clientConn, 100)

	c.incrementDownloadedBytes(10)
	if c.isDownloadBytesLimitExceeded() {
		t.Errorf("expected download limit not to be exceeded (limit: %d, used: %d), but it was", c.maxBytes, c.uploadedBytes)
	}

	c.incrementDownloadedBytes(89)
	if c.isDownloadBytesLimitExceeded() {
		t.Errorf("expected download limit not to be exceeded after adding 99 total bytes (limit: %d, used: %d)", c.maxBytes, c.uploadedBytes)
	}

	c.incrementDownloadedBytes(1)
	if !c.isDownloadBytesLimitExceeded() {
		t.Errorf("expected download limit to be exceeded after adding 100 total bytes (limit: %d, used: %d)", c.maxBytes, c.uploadedBytes)
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
			c.incrementUploadedBytes(10)
		}()
	}
	wg.Wait()
	if c.uploadedBytes != 1000 {
		t.Errorf("expected sentBytes to be 1000, got %d", c.uploadedBytes)
	}
}
