package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWsHandler(t *testing.T) {
	// Create a test logger
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
		Formatter:       log.TextFormatter,
	})

	// Create test config
	cfg := &APIConfig{
		Logger: logger,
		Hub:    NewHub(),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(cfg.WsHandler))
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Test successful connection
	t.Run("Successful Connection", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err, "should connect to websocket without error")
		defer conn.Close()

		// Verify client was added to hub
		cfg.Hub.mu.Lock()
		assert.Equal(t, 1, len(cfg.Hub.clients), "should have exactly one client in hub")
		cfg.Hub.mu.Unlock()
	})

	// Test message broadcasting
	t.Run("Message Broadcasting", func(t *testing.T) {
		// Create two test clients
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err, "should connect first client without error")
		defer conn1.Close()

		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err, "should connect second client without error")
		defer conn2.Close()

		//  Wait and Verify both clients were added to hub
		time.Sleep(100 * time.Millisecond) // Allow some time for connections to be processed
		cfg.Hub.mu.Lock()
		require.Equal(t, 2, len(cfg.Hub.clients), "should have exactly two clients in hub")
		cfg.Hub.mu.Unlock()

		// Set up message receivers
		messageChan1 := make(chan []byte)
		messageChan2 := make(chan []byte)

		go func() {
			_, message, err := conn1.ReadMessage()
			assert.NoError(t, err, "should read message from client 1 without error")
			messageChan1 <- message
		}()

		go func() {
			_, message, err := conn2.ReadMessage()
			assert.NoError(t, err, "should read message from client 2 without error")
			messageChan2 <- message
		}()

		// Broadcast a test message
		testMessage := []byte("test message")
		cfg.Hub.Broadcast(testMessage)

		// Wait for messages with timeout
		select {
		case msg1 := <-messageChan1:
			assert.Equal(t, testMessage, msg1, "client 1 should receive the correct message")
		case <-time.After(time.Second):
			t.Error("timeout waiting for message to client 1")
		}

		select {
		case msg2 := <-messageChan2:
			assert.Equal(t, testMessage, msg2, "client 2 should receive the correct message")
		case <-time.After(time.Second):
			t.Error("timeout waiting for message to client 2")
		}
	})

	// Test client disconnection
	t.Run("Client Disconnection", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err, "should connect to websocket without error")

		// Close the connection
		conn.Close()

		// Wait a bit for the disconnection to be processed
		time.Sleep(100 * time.Millisecond)

		// Verify client was removed from hub
		cfg.Hub.mu.Lock()
		assert.Equal(t, 0, len(cfg.Hub.clients), "should have no clients in hub after disconnection")
		cfg.Hub.mu.Unlock()
	})
}
