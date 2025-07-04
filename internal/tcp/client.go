package tcp

import (
	"net"
	"sync"
	"sync/atomic"
)

type client struct {
	conn            net.Conn
	uploadedBytes   int64
	downloadedBytes int64
	maxBytes        int64
	sendCh          chan []byte
	once            sync.Once
}

func newClient(conn net.Conn, maxBytes int64) *client {
	return &client{
		conn:     conn,
		maxBytes: maxBytes,
		sendCh:   make(chan []byte, 10),
	}
}

func (c *client) isDownloadBytesLimitExceeded() bool {
	return atomic.LoadInt64(&c.downloadedBytes) >= c.maxBytes
}

func (c *client) isUploadedBytesLimitExceeded() bool {
	return atomic.LoadInt64(&c.uploadedBytes) >= c.maxBytes
}

func (c *client) incrementUploadedBytes(size int) {
	atomic.AddInt64(&c.uploadedBytes, int64(size))
}

func (c *client) incrementDownloadedBytes(size int) {
	atomic.AddInt64(&c.downloadedBytes, int64(size))
}
