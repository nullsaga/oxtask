package tcp

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
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

func TestDownloadedMaxBytesNotReached(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	sender := connectClient(t, addr)
	defer sender.Close()

	receiver := connectClient(t, addr)
	defer receiver.Close()
	expectedMsg := "hello world\n"

	_, err := sender.Write([]byte(expectedMsg))
	if err != nil {
		t.Fatalf("sender failed to write: %v", err)
	}

	r := bufio.NewReader(receiver)
	recvMsg, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("receiver read failed: %v", err)
	}

	if recvMsg != expectedMsg {
		t.Fatalf("receiver got %q, want %q", recvMsg, expectedMsg)
	}
}

func TestDownloadedMaxBytesReached(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	sender := connectClient(t, addr)
	defer sender.Close()

	receiver := connectClient(t, addr)
	receiver.SetReadDeadline(time.Now().Add(1 * time.Second))
	defer receiver.Close()

	_, err := sender.Write([]byte(strings.Repeat("t", 101) + "\n"))
	if err != nil {
		t.Fatalf("sender failed to write: %v", err)
	}

	r := bufio.NewReader(receiver)
	_, err = r.ReadString('\n')
	var netErr net.Error
	if err != nil {
		if !errors.As(err, &netErr) && netErr.Timeout() {
			t.Fatalf("read failed unexpectedly: %v", err)
		}
	}
}

func TestUploadedMaxBytesNotReached(t *testing.T) {
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

func TestUploadedMaxBytesReached(t *testing.T) {
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
