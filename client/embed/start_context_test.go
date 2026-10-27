package embed

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestStartCancellationInterruptsAuthentication(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		close(accepted)
		buffer := make([]byte, 4096)
		for {
			if _, err := conn.Read(buffer); err != nil {
				return
			}
		}
	}()
	client, err := New(Options{ManagementURL: "https://" + listener.Addr().String(), SetupKey: "unused-test-key", DeviceName: "cancellation-test"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- client.Start(ctx) }()
	select {
	case <-accepted:
	case <-time.After(5 * time.Second):
		t.Fatal("authentication never reached the test listener")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled startup succeeded")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("startup ignored cancellation during authentication")
	}
}
