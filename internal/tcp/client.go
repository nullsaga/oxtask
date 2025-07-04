package tcp

import (
	"net"
	"sync"
	"sync/atomic"
)

type client struct {
	conn      net.Conn
	sentBytes int64
	maxBytes  int64
	sendCh    chan []byte
	once      sync.Once
}

func newClient(conn net.Conn, maxBytes int64) *client {
	return &client{
		conn:     conn,
		maxBytes: maxBytes,
		sendCh:   make(chan []byte, 10),
	}
}

func (c *client) isBandwidthLimitExceeded() bool {
	return atomic.LoadInt64(&c.sentBytes) >= c.maxBytes
}

func (c *client) incrementSentBytes(size int) {
	atomic.AddInt64(&c.sentBytes, int64(size))
}
