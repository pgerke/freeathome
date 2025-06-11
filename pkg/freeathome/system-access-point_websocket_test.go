package freeathome

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pgerke/freeathome/pkg/models"
)

const testMessageValid = "valid message"

// TestSystemAccessPointWebSocketMessageHandler tests the webSocketMessageHandler method of SystemAccessPoint.
func TestSystemAccessPointWebSocketMessageHandler(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	defer sysAp.waitGroup.Wait()
	sysAp.webSocketMessageChannel = make(chan []byte, 10)

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
		close(sysAp.webSocketMessageChannel)
		sysAp.webSocketMessageChannel = nil
	}()

	// Start the handler
	sysAp.webSocketMessageHandler()

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

func TestSystemAccessPointWebSocketMessageHandlerMissingChannel(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	defer sysAp.waitGroup.Wait()
	sysAp.webSocketMessageChannel = nil

	// Start the handler
	sysAp.webSocketMessageHandler()

	// Check the log output
	logOutput := buf.String()

	if !strings.Contains(logOutput, "webSocketMessageChannel is nil") {
		t.Errorf("Expected log output to contain 'webSocketMessageChannel is nil', got: %s", logOutput)
	}
}

// TestSystemAccessPointConnectWebSocketSuccess tests the successful connection of the WebSocket.
func TestSystemAccessPointConnectWebSocketSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, _, records := setup(t, false)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	var conn *websocket.Conn
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade WebSocket: %v", err)
		}
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
					_ = conn.Close()
				}
			}
		}
	}()

	// Run ConnectWebSocket in a separate goroutine
	sysAp.ConnectWebSocket(ctx, 1*time.Hour)
}

// TestSystemAccessPointConnectWebSocketContextCancelled tests the behavior when the context is cancelled.
func TestSystemAccessPointConnectWebSocketContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sysAp, _, records := setup(t, false)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	var conn *websocket.Conn
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade WebSocket: %v", err)
		}
	}))
	defer server.Close()

	sysAp.hostName = strings.TrimPrefix(server.URL, "http://")

	wg := sync.WaitGroup{}
	wg.Add(2)

	innerCtx, innerCancel := context.WithCancel(context.TODO())
	go func() {
		for {
			select {
			case <-innerCtx.Done():
				return
			case record := <-records:
				// Cancel the context when the web socket is connected successfully
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, "web socket connected successfully") {
					cancel()
					_ = conn.WriteMessage(websocket.TextMessage, []byte("test"))
					break
				}
				// Send one done when the message handler is stopped
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, "webSocketMessageChannel closed, stopping message handler") {
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
		sysAp.ConnectWebSocket(ctx, 1*time.Hour)
	}()

	// Wait for the expected records to be processed
	wg.Wait()
	// Cancel the inner context to stop the record channel reader
	innerCancel()
}

// TestSystemAccessPointConnectWebSocketFailure tests the behavior when the WebSocket connection fails.
func TestSystemAccessPointConnectWebSocketFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, false)
	defer sysAp.waitGroup.Wait()

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
		sysAp.ConnectWebSocket(ctx, 1*time.Hour)
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "failed to connect to web socket") {
		t.Errorf("Expected log output to contain 'failed to connect to web socket', got: %s", logOutput)
	}
}

// TestSystemAccessPointwebSocketMessageLoopTextMessage tests the webSocketMessageLoop method for text messages.
func TestSystemAccessPointwebSocketMessageLoopTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, true)
	sysAp.webSocketMessageChannel = make(chan []byte, 10)
	sysAp.messageReceivedChannel = make(chan struct{}, 1)

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte(testMessageValid),
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	go func() {
		err := sysAp.webSocketMessageLoop(ctx, conn)
		if err == nil {
			t.Error(expectedErrorGotNil)
		}
	}()

	// Wait for the context to be done
	message := <-sysAp.webSocketMessageChannel
	cancel()
	<-ctx.Done()
	close(sysAp.webSocketMessageChannel)
	sysAp.webSocketMessageChannel = nil
	close(sysAp.messageReceivedChannel)
	sysAp.messageReceivedChannel = nil
	sysAp.waitGroup.Wait()

	// Check if the message is valid
	if string(message) != testMessageValid {
		t.Errorf("Expected message 'valid message', got: %s", string(message))
	}

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "received text message from web socket") {
		t.Errorf("Expected log output to contain 'received text message from web socket', got: %s", logOutput)
	}
}

// TestSystemAccessPointwebSocketMessageLoopNonTextMessage tests the webSocketMessageLoop method for non-text messages.
func TestSystemAccessPointwebSocketMessageLoopNonTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, true)
	sysAp.webSocketMessageChannel = make(chan []byte, 10)
	sysAp.messageReceivedChannel = make(chan struct{}, 1)
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
			t.Error(expectedErrorGotNil)
		}
	}()

	// Wait for the context to be done
	<-ctx.Done()
	close(sysAp.webSocketMessageChannel)
	sysAp.webSocketMessageChannel = nil
	close(sysAp.messageReceivedChannel)
	sysAp.messageReceivedChannel = nil
	sysAp.waitGroup.Wait()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "received non-text message from web socket") {
		t.Errorf("Expected log output to contain 'received non-text message from web socket', got: %s", logOutput)
	}
}

func TestSystemAccessPointwebSocketMessageLoopMissingChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysAp, buf, _ := setup(t, true)
	sysAp.webSocketMessageChannel = nil
	sysAp.messageReceivedChannel = make(chan struct{}, 1)

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte(testMessageValid),
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	err := sysAp.webSocketMessageLoop(ctx, conn)
	if err == nil {
		t.Error(expectedErrorGotNil)
	}
	// Check if the error is due to the missing channel
	if !strings.Contains(err.Error(), "a connection channel is nil, cannot start message loop") {
		t.Errorf("Expected error 'a connection channel is nil, cannot start message loop', got: %v", err)
	}

	// Wait for the context to be done
	cancel()
	close(sysAp.messageReceivedChannel)
	sysAp.messageReceivedChannel = nil
	sysAp.waitGroup.Wait()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "a connection channel is nil, cannot start message loop") {
		t.Errorf("Expected log output to contain 'a connection channel is nil, cannot start message loop', got: %s", logOutput)
	}
}

func TestSystemAccessPointwebSocketKeepaliveLoopMissingChannel(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	sysAp.messageReceivedChannel = nil

	// Mock a WebSocket connection
	conn := &MockConn{
		err: nil,
	}

	// Run the keepalive loop in a separate goroutine
	sysAp.webSocketKeepaliveLoop(conn, 30*time.Second)

	// Wait for the context to be done
	sysAp.waitGroup.Wait()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "essageReceivedChannel is nil, cannot start keepalive loop") {
		t.Errorf("Expected log output to contain 'essageReceivedChannel is nil, cannot start keepalive loop', got: %s", logOutput)
	}
}

func TestSystemAccessPointwebSocketKeepaliveLoopSendPing(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	sysAp.messageReceivedChannel = make(chan struct{}, 1)

	// Mock a WebSocket connection
	conn := &MockConn{
		err: errors.New("test error"),
		writeMessages: []struct {
			messageType int
			data        []byte
			deadline    time.Time
		}{},
	}

	// Run the keepalive loop in a separate goroutine
	go func() {
		sysAp.webSocketKeepaliveLoop(conn, 250*time.Millisecond)
	}()

	time.Sleep(150 * time.Millisecond)
	if len(conn.writeMessages) != 0 {
		t.Errorf("Expected write message count to be 0, got: %d", len(conn.writeMessages))
	}
	time.Sleep(150 * time.Millisecond)

	// Wait for the context to be done
	close(sysAp.messageReceivedChannel)
	sysAp.messageReceivedChannel = nil
	sysAp.waitGroup.Wait()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "keepalive timer expired, sending ping") {
		t.Errorf("Expected log output to contain 'keepalive timer expired, sending ping', got: %s", logOutput)
	}

	// Check if the ping message was sent
	if len(conn.writeMessages) != 1 {
		t.Errorf("Expected write message count to be 1, got: %d", len(conn.writeMessages))
	}
}

type MockConn struct {
	messageRead   bool
	messageType   int
	r             []byte
	err           error
	writeMessages []struct {
		messageType int
		data        []byte
		deadline    time.Time
	}
}

func (m *MockConn) ReadMessage() (int, []byte, error) {
	if m.messageRead {
		return -1, nil, fmt.Errorf("no more messages")
	}

	m.messageRead = true
	return m.messageType, m.r, m.err
}

func (m *MockConn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	m.writeMessages = append(m.writeMessages, struct {
		messageType int
		data        []byte
		deadline    time.Time
	}{
		messageType: messageType,
		data:        data,
		deadline:    deadline,
	})
	return m.err
}
