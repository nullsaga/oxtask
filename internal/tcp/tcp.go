package tcp

import (
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
