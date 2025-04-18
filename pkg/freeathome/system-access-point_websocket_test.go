package freeathome

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/pgerke/freeathome/pkg/models"
)

// TestSystemAccessPoint_WebSocketMessageHandler tests the webSocketMessageHandler method of SystemAccessPoint.
func TestSystemAccessPoint_WebSocketMessageHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, true)

	// Mock a valid WebSocketMessage
	validMessage := models.WebSocketMessage{
		models.EmptyUUID: models.Message{
			Datapoints: map[string]string{
				"ABB7F595EC47/ch0000/odp0000": "1",
			},
		},
	}
	validMessageBytes, _ := json.Marshal(validMessage)

	// Mock an invalid WebSocketMessage
	invalidMessage := []byte(`invalid json`)

	// Mock a WebSocketMessage with no datapoints
	emptyMessage := models.WebSocketMessage{
		models.EmptyUUID: models.Message{
			Datapoints: map[string]string{},
		},
	}
	emptyMessageBytes, _ := json.Marshal(emptyMessage)

	// Mock a WebSocketMessage with invalid datapoint format
	invalidFormatMessage := models.WebSocketMessage{
		models.EmptyUUID: models.Message{
			Datapoints: map[string]string{
				"Test123": "1",
			},
		},
	}
	invalidFormatMessageBytes, _ := json.Marshal(invalidFormatMessage)

	// Send messages to the WebSocketMessageChannel
	var wg sync.WaitGroup
	wg.Add(4)
	sysAp.onMessageHandled = wg.Done
	go func() {
		sysAp.webSocketMessageChannel <- validMessageBytes
		sysAp.webSocketMessageChannel <- invalidMessage
		sysAp.webSocketMessageChannel <- emptyMessageBytes
		sysAp.webSocketMessageChannel <- invalidFormatMessageBytes
		wg.Wait()
		cancel() // Stop the handler after processing the messages
	}()

	// Start the handler
	sysAp.webSocketMessageHandler(ctx)
	cancel()

	// Check the log output
	logOutput := buf.String()

	// Verify valid message processing
	if !strings.Contains(logOutput, "data point update") {
		t.Errorf("Expected log output to contain 'data point update', got: %s", logOutput)
	}

	// Verify invalid message handling
	if !strings.Contains(logOutput, "failed to unmarshal message") {
		t.Errorf("Expected log output to contain 'failed to unmarshal message', got: %s", logOutput)
	}

	// Verify empty message handling
	if !strings.Contains(logOutput, "web socket message has no datapoints") {
		t.Errorf("Expected log output to contain 'web socket message has no datapoints', got: %s", logOutput)
	}

	// Verify invalid format message handling
	if !strings.Contains(logOutput, "Ignored datapoint with invalid key format") {
		t.Errorf("Expected log output to contain 'Ignored datapoint with invalid key format', got: %s", logOutput)
	}
}

// TestSystemAccessPoint_ConnectWebSocket_Success tests the successful connection of the WebSocket.
func TestSystemAccessPoint_ConnectWebSocket_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, _, records := setup(t, false)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade WebSocket: %v", err)
		}
		defer func() {
			_ = conn.Close()
		}()
	}))
	defer server.Close()

	sysAp.hostName = strings.TrimPrefix(server.URL, "http://")

	// Wait for the expected record in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case record := <-records:
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, "web socket connected successfully") {
					cancel()
				}
			}
		}
	}()

	// Run ConnectWebSocket in a separate goroutine
	go func() {
		sysAp.ConnectWebSocket(ctx)
	}()

	<-ctx.Done()
}

// TestSystemAccessPoint_ConnectWebSocket_ContextCancelled tests the behavior when the context is cancelled.
func TestSystemAccessPoint_ConnectWebSocket_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sysAp, _, records := setup(t, false)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade WebSocket: %v", err)
		}
		defer func() {
			_ = conn.Close()
		}()
	}))
	defer server.Close()

	sysAp.hostName = strings.TrimPrefix(server.URL, "http://")

	wg := sync.WaitGroup{}
	wg.Add(2)

	innerCtx, innerCancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-innerCtx.Done():
				return
			case record := <-records:
				// Cancel the context when the web socket is connected successfully
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, "web socket connected successfully") {
					cancel()
					break
				}
				// Send one done when the message handler is stopped
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, "context cancelled, stopping message handler") {
					wg.Done()
					break
				}
				// Send one done when the web socket connection is stopped
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, "context cancelled, stopping web socket connection attempts") {
					wg.Done()
				}
			}
		}
	}()

	// Run ConnectWebSocket in a separate goroutine
	go func() {
		sysAp.ConnectWebSocket(ctx)
	}()

	// Wait for the expected records to be processed
	wg.Wait()
	// Cancel the inner context to stop the record channel reader
	innerCancel()
}

// TestSystemAccessPoint_ConnectWebSocket_Failure tests the behavior when the WebSocket connection fails.
func TestSystemAccessPoint_ConnectWebSocket_Failure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, false)

	// Set an invalid host name to simulate connection failure
	sysAp.hostName = "invalid-host"

	// set up the error handler
	sysAp.onError = func(err error) {
		if strings.Contains(err.Error(), "lookup invalid-host") {
			cancel()
		} else {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Run ConnectWebSocket in a separate goroutine
	go func() {
		sysAp.ConnectWebSocket(ctx)
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "failed to connect to web socket") {
		t.Errorf("Expected log output to contain 'failed to connect to web socket', got: %s", logOutput)
	}
}

// TestSystemAccessPoint_webSocketMessageLoop_TextMessage tests the webSocketMessageLoop method for text messages.
func TestSystemAccessPoint_webSocketMessageLoop_TextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, true)

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte("valid message"),
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	go func() {
		err := sysAp.webSocketMessageLoop(ctx, conn)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}()

	// Wait for the context to be done
	message := <-sysAp.webSocketMessageChannel
	cancel()
	<-ctx.Done()

	// Check if the message is valid
	if string(message) != "valid message" {
		t.Errorf("Expected message 'valid message', got: %s", string(message))
	}

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "received text message from web socket") {
		t.Errorf("Expected log output to contain 'received text message from web socket', got: %s", logOutput)
	}
}

// TestSystemAccessPoint_webSocketMessageLoop_NonTextMessage tests the webSocketMessageLoop method for non-text messages.
func TestSystemAccessPoint_webSocketMessageLoop_NonTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, true)
	sysAp.onError = func(err error) {
		if strings.Contains(err.Error(), "no more messages") {
			cancel()
		} else {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Mock a non-text message
	nonTextMessage := []byte{0x00, 0x01, 0x02}

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.BinaryMessage,
		r:           nonTextMessage,
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	go func() {
		err := sysAp.webSocketMessageLoop(ctx, conn)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}()

	// Wait for the context to be done
	<-ctx.Done()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "received non-text message from web socket") {
		t.Errorf("Expected log output to contain 'received non-text message from web socket', got: %s", logOutput)
	}
}

type MockConn struct {
	messageRead bool
	messageType int
	r           []byte
	err         error
}

func (m *MockConn) ReadMessage() (int, []byte, error) {
	if m.messageRead {
		return -1, nil, fmt.Errorf("no more messages")
	}

	m.messageRead = true
	return m.messageType, m.r, m.err
}
