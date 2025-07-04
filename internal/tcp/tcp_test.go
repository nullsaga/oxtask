package tcp

import (
	"bufio"
	"net"
	"strings"
	"testing"
)

func startTestServer(t *testing.T) (string, func()) {
	t.Helper()
	srv, err := NewServer("localhost:0")
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	go func() {
		if err = srv.Serve(); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()

	cleanup := func() {
		if err = srv.Shutdown(); err != nil {
			t.Errorf("server shutdown error: %v", err)
		}
	}

	return srv.listener.Addr().String(), cleanup
}

func connectClient(t *testing.T, addr string) net.Conn {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("client failed to connect: %v", err)
	}

	return c
}

func TestCanBroadcast(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	client1 := connectClient(t, addr)
	client2 := connectClient(t, addr)
	defer client1.Close()
	defer client2.Close()

	msg := "hello world\n"
	_, err := client1.Write([]byte(msg))
	if err != nil {
		t.Fatalf("client1 failed to write: %v", err)
	}

	r := bufio.NewReader(client2)
	recvMsg, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("client2 failed to read: %v", err)
	}

	if recvMsg != msg {
		t.Fatalf("client2 got %q, want %q", recvMsg, msg)
	}
}

func TestMaxBytesNotReached(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn := connectClient(t, addr)
	defer conn.Close()

	_, err := conn.Write([]byte(strings.Repeat("x", 50) + "\n"))
	if err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err = conn.Write([]byte("x\n"))
		if err != nil {
			t.Fatalf("write %d failed: %v", i+1, err)
		}
	}

	_, err = conn.Write([]byte("final write\n"))
	if err != nil {
		t.Fatalf("final write failed: %v", err)
	}
}

func TestMaxBytesReached(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn := connectClient(t, addr)
	defer conn.Close()

	_, err := conn.Write([]byte(strings.Repeat("x", 101) + "\n"))
	if err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("expected disconnect message, but got error: %v", err)
	}

	expected := "Disconnected due to exceeding uploaded bytes limit\n"
	if msg != expected {
		t.Fatalf("expected message %q, got %q", expected, msg)
	}

	t.Logf("received disconnect message: %q", msg)
}
