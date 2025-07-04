package tcp

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	listener net.Listener
	mutex    sync.Mutex
	clients  map[*client]struct{}
}

func NewServer(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: ln,
		mutex:    sync.Mutex{},
		clients:  make(map[*client]struct{}),
	}, nil
}

func (s *Server) Serve() error {
	defer s.listener.Close()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		c := newClient(conn, 100)
		s.mutex.Lock()
		s.clients[c] = struct{}{}
		s.mutex.Unlock()

		go s.handleClientConn(c)
	}
}

func (s *Server) handleClientConn(c *client) {
	defer c.conn.Close()
	r := bufio.NewReader(c.conn)
	go s.write(c)

	for {
		msg, err := r.ReadBytes('\n')
		if err != nil {
			break
		}

		c.incrementUploadedBytes(len(msg))
		if c.isUploadedBytesLimitExceeded() {
			_, _ = c.conn.Write([]byte("Disconnected due to exceeding uploaded bytes limit\n"))
			break
		}
		s.broadcast(c, msg)
	}

	s.removeClient(c)
}

func (s *Server) write(c *client) {
	for msg := range c.sendCh {
		n, err := c.conn.Write(msg)
		if err != nil {
			fmt.Printf("write error to %s: %v\n", c.conn.RemoteAddr(), err)
			break
		}

		c.incrementDownloadedBytes(n)
		if c.isDownloadBytesLimitExceeded() {
			_, _ = c.conn.Write([]byte("Disconnected due to exceeding downloaded bytes limit\n"))
			break
		}
	}

	s.removeClient(c)
}

func (s *Server) broadcast(sender *client, msg []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for c := range s.clients {
		if c == sender {
			continue
		}

		select {
		case c.sendCh <- msg:
		default:
			fmt.Printf("client %s too slow, disconnecting\n", c.conn.RemoteAddr())
			s.removeClient(c)
		}
	}
}

func (s *Server) removeClient(c *client) {
	c.once.Do(func() {
		s.mutex.Lock()
		delete(s.clients, c)
		s.mutex.Unlock()
		close(c.sendCh)
		c.conn.Close()
	})
}

func (s *Server) Shutdown() error {
	return s.listener.Close()
}
