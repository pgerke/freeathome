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
	ws, buf, _ := setupSysApWebSocket(t, true, false)
	defer ws.waitGroup.Wait()

	webSocketMessageChannel := make(chan []byte, 10)

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
	ws.onMessageHandled = wg.Done
	go func() {
		webSocketMessageChannel <- validMessageBytes
		webSocketMessageChannel <- invalidMessage
		webSocketMessageChannel <- emptyMessageBytes
		webSocketMessageChannel <- invalidFormatMessageBytes
		wg.Wait()
		close(webSocketMessageChannel)
		webSocketMessageChannel = nil
	}()

	// Start the handler
	ws.webSocketMessageHandler(webSocketMessageChannel)

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
	ws, buf, _ := setupSysApWebSocket(t, true, false)
	defer ws.waitGroup.Wait()

	// Start the handler
	ws.webSocketMessageHandler(nil)

	// Check the log output
	logOutput := buf.String()

	if !strings.Contains(logOutput, "webSocketMessageChannel is nil") {
		t.Errorf("Expected log output to contain 'webSocketMessageChannel is nil', got: %s", logOutput)
	}
}

// TestSystemAccessPointConnectWebSocketSuccess tests the successful connection of the WebSocket.
func TestSystemAccessPointConnectWebSocketSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sysAp, _, records := setupSysAp(t, false, false)

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

	sysAp.config.Hostname = strings.TrimPrefix(server.URL, "http://")

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
	err := sysAp.ConnectWebSocket(ctx, 1*time.Hour)
	if err != nil && err != context.Canceled {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestSystemAccessPointConnectWebSocketSkipTlsVerify tests the successful connection of the WebSocket with skip TLS verify.
func TestSystemAccessPointConnectWebSocketSkipTlsVerify(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sysAp, buf, records := setupSysAp(t, true, true)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	var conn *websocket.Conn
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade WebSocket: %v", err)
		}
	}))
	defer server.Close()

	sysAp.config.Hostname = strings.TrimPrefix(server.URL, "https://")

	// Wait for the expected record in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Check the log output
				logOutput := buf.String()
				if !strings.Contains(logOutput, "this is not recommended") {
					t.Errorf("Expected log output to contain 'this is not recommended', got: %s", logOutput)
				}
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
	err := sysAp.ConnectWebSocket(ctx, 1*time.Hour)
	if err != nil && err != context.Canceled {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestSystemAccessPointConnectWebSocketContextCancelled tests the behavior when the context is cancelled.
func TestSystemAccessPointConnectWebSocketContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	sysAp, _, records := setupSysAp(t, false, false)

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

	sysAp.config.Hostname = strings.TrimPrefix(server.URL, "http://")

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
		err := sysAp.ConnectWebSocket(ctx, 1*time.Hour)
		if err != nil && err != context.Canceled {
			t.Errorf("Expected no error, got: %v", err)
		}
	}()

	// Wait for the expected records to be processed
	wg.Wait()
	// Cancel the inner context to stop the record channel reader
	innerCancel()
}

// TestSystemAccessPointConnectWebSocketFailure tests the behavior when the WebSocket connection fails.
func TestSystemAccessPointConnectWebSocketFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sysAp, buf, _ := setupSysAp(t, false, false)

	// Set an invalid host name to simulate connection failure
	sysAp.config.Hostname = "invalid-host"

	// set up the error handler
	sysAp.onError = func(err error) {
		if strings.Contains(err.Error(), "lookup invalid-host") {
			cancel()
		} else {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// TODO: Add reconnection attempts test when implemented
	// // Get the initial reconnection attempts
	// reconnectionAttempts := ws.sysAp.GetReconnectionAttempts()
	// if reconnectionAttempts != 0 {
	// 	t.Errorf("Expected reconnection attempts to be 0, got %d", reconnectionAttempts)
	// }

	// Run ConnectWebSocket in a separate goroutine
	go func() {
		err := sysAp.ConnectWebSocket(ctx, 1*time.Hour)
		if err != nil && err != context.Canceled {
			t.Errorf("Expected no error, got: %v", err)
		}
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// reconnectionAttempts = sysAp.GetReconnectionAttempts()
	// if reconnectionAttempts != 1 {
	// 	t.Errorf("Expected reconnection attempts to be 1, got %d", reconnectionAttempts)
	// }

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "failed to connect to web socket") {
		t.Errorf("Expected log output to contain 'failed to connect to web socket', got: %s", logOutput)
	}
}

// func TestSystemAccessPointConnectWebSocketMaxReconnectionAttempts(t *testing.T) {
// 	sysAp, buf, _ := setupSysAp(t, false, false)
// 	defer sysAp.waitGroup.Wait()

// 	// Set the max reconnection attempts to 2
// 	sysAp.SetMaxReconnectionAttempts(2)

// 	// Set an invalid host name to simulate connection failure
// 	sysAp.config.Hostname = "invalid-host"

// 	// set up the error handler
// 	errorCount := 0
// 	sysAp.onError = func(err error) {
// 		errorCount++
// 	}

// 	// Get the initial reconnection attempts
// 	reconnectionAttempts := sysAp.GetReconnectionAttempts()
// 	if reconnectionAttempts != 0 {
// 		t.Errorf("Expected reconnection attempts to be 0, got %d", reconnectionAttempts)
// 	}

// 	// Run ConnectWebSocket
// 	err := sysAp.ConnectWebSocket(t.Context(), 1*time.Hour)

// 	// Verify error
// 	if err == nil || err.Error() != "maximum reconnection attempts exceeded" {
// 		t.Errorf("Expected error 'maximum reconnection attempts exceeded', got: %v", err)
// 	}

// 	// Verify the reconnection attempts
// 	reconnectionAttempts = sysAp.GetReconnectionAttempts()
// 	if reconnectionAttempts != 2 {
// 		t.Errorf("Expected reconnection attempts to be 2, got %d", reconnectionAttempts)
// 	}

// 	// Verify the error count
// 	if errorCount != 2 {
// 		t.Errorf("Expected error count to be 2, got %d", errorCount)
// 	}

// 	// Check the log output
// 	logOutput := buf.String()
// 	if !strings.Contains(logOutput, "maximum reconnection attempts exceeded") {
// 		t.Errorf("Expected log output to contain 'maximum reconnection attempts exceeded', got: %s", logOutput)
// 	}
// }

// TestSystemAccessPointWebSocketMessageLoopTextMessage tests the webSocketMessageLoop method for text messages.
func TestSystemAccessPointWebSocketMessageLoopTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ws, buf, _ := setupSysApWebSocket(t, true, false)
	webSocketMessageChannel := make(chan []byte, 10)
	messageReceivedChannel := make(chan struct{}, 1)

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte(testMessageValid),
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	go func() {
		err := ws.webSocketMessageLoop(ctx, messageReceivedChannel, webSocketMessageChannel, conn)
		if err == nil {
			t.Error(expectedErrorGotNil)
		}
	}()

	// Wait for the context to be done
	message := <-webSocketMessageChannel
	cancel()
	<-ctx.Done()
	close(webSocketMessageChannel)
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

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

// TestSystemAccessPointWebSocketMessageLoopNonTextMessage tests the webSocketMessageLoop method for non-text messages.
func TestSystemAccessPointWebSocketMessageLoopNonTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ws, buf, _ := setupSysApWebSocket(t, true, false)
	webSocketMessageChannel := make(chan []byte, 10)
	messageReceivedChannel := make(chan struct{}, 1)
	ws.sysAp.onError = func(err error) {
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
		err := ws.webSocketMessageLoop(ctx, messageReceivedChannel, webSocketMessageChannel, conn)
		if err == nil {
			t.Error(expectedErrorGotNil)
		}
	}()

	// Wait for the context to be done
	<-ctx.Done()
	close(webSocketMessageChannel)
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "received non-text message from web socket") {
		t.Errorf("Expected log output to contain 'received non-text message from web socket', got: %s", logOutput)
	}
}

// TestSystemAccessPointWebSocketMessageLoopMissingChannel tests the webSocketMessageLoop method for missing channels.
func TestSystemAccessPointWebSocketMessageLoopMissingChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ws, buf, _ := setupSysApWebSocket(t, true, false)
	messageReceivedChannel := make(chan struct{}, 1)

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte(testMessageValid),
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	err := ws.webSocketMessageLoop(ctx, messageReceivedChannel, nil, conn)
	if err == nil {
		t.Error(expectedErrorGotNil)
	}
	// Check if the error is due to the missing channel
	if !strings.Contains(err.Error(), "a connection channel is nil, cannot start message loop") {
		t.Errorf("Expected error 'a connection channel is nil, cannot start message loop', got: %v", err)
	}

	// Wait for the context to be done
	cancel()
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "a connection channel is nil, cannot start message loop") {
		t.Errorf("Expected log output to contain 'a connection channel is nil, cannot start message loop', got: %s", logOutput)
	}
}

// TestSystemAccessPointwebSocketKeepaliveLoopMissingChannel tests the webSocketKeepaliveLoop method for missing channels.
func TestSystemAccessPointwebSocketKeepaliveLoopMissingChannel(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)

	// Mock a WebSocket connection
	conn := &MockConn{
		err: nil,
	}

	// Run the keepalive loop in a separate goroutine
	ws.webSocketKeepaliveLoop(nil, conn, 30*time.Second)

	// Wait for the context to be done
	ws.waitGroup.Wait()

	// Check the log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "essageReceivedChannel is nil, cannot start keepalive loop") {
		t.Errorf("Expected log output to contain 'essageReceivedChannel is nil, cannot start keepalive loop', got: %s", logOutput)
	}
}

// TestSystemAccessPointwebSocketKeepaliveLoopSendPing tests the webSocketKeepaliveLoop method for sending a ping message.
func TestSystemAccessPointwebSocketKeepaliveLoopSendPing(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)
	messageReceivedChannel := make(chan struct{}, 1)

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
		ws.webSocketKeepaliveLoop(messageReceivedChannel, conn, 250*time.Millisecond)
	}()

	time.Sleep(150 * time.Millisecond)
	if len(conn.writeMessages) != 0 {
		t.Errorf("Expected write message count to be 0, got: %d", len(conn.writeMessages))
	}
	time.Sleep(150 * time.Millisecond)

	// Wait for the context to be done
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

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

// TestSystemAccessPointGetWebSocketUrlWithoutTls tests the getWebSocketUrl method of SystemAccessPointWebSocket without TLS.
func TestSystemAccessPointGetWsUrlWithoutTls(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, false, false)

	actual := ws.getWebSocketUrl()
	expected := "ws://localhost/fhapi/v1/api/ws"

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}

// TestSystemAccessPointGetWebSocketUrlWithTls tests the getWebSocketUrl method of SystemAccessPointWebSocket with TLS.
func TestSystemAccessPointGetWsUrlWithTls(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)

	actual := ws.getWebSocketUrl()
	expected := "wss://localhost/fhapi/v1/api/ws"

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}

// // TestSystemAccessPointGetSetMaxReconnectionAttempts tests the GetMaxReconnectionAttempts and SetMaxReconnectionAttempts methods of SystemAccessPoint.
// func TestSystemAccessPointGetSetMaxReconnectionAttempts(t *testing.T) {
// 	sysAp, _, _ := setupSysAp(t, true, false)

// 	// Test initial value
// 	expected := 3
// 	actual := sysAp.GetMaxReconnectionAttempts()
// 	if actual != expected {
// 		t.Errorf("Expected max reconnection attempts to be %d, got %d", expected, actual)
// 	}

// 	// Test setting a new value
// 	expected = 5
// 	sysAp.SetMaxReconnectionAttempts(expected)
// 	actual = sysAp.GetMaxReconnectionAttempts()
// 	if actual != expected {
// 		t.Errorf("Expected max reconnection attempts to be %d, got %d", expected, actual)
// 	}
// }

// // TestSystemAccessPointGetSetExponentialBackoffEnabled tests the GetExponentialBackoffEnabled and SetExponentialBackoffEnabled methods of SystemAccessPoint.
// func TestSystemAccessPointGetSetExponentialBackoffEnabled(t *testing.T) {
// 	sysAp, _, _ := setupSysAp(t, true, false)

// 	// Test initial value
// 	expected := true
// 	actual := sysAp.GetExponentialBackoffEnabled()
// 	if actual != expected {
// 		t.Errorf("Expected exponential backoff enabled to be %v, got %v", expected, actual)
// 	}

// 	// Test setting a new value
// 	expected = false
// 	sysAp.SetExponentialBackoffEnabled(expected)
// 	actual = sysAp.GetExponentialBackoffEnabled()
// 	if actual != expected {
// 		t.Errorf("Expected exponential backoff enabled to be %v, got %v", expected, actual)
// 	}
// }

// // TestSystemAccessPointCalculateBackoffDuration tests the calculateBackoffDuration method of SystemAccessPoint.
// func TestSystemAccessPointCalculateBackoffDuration(t *testing.T) {
// 	sysAp, _, _ := setupSysAp(t, true, false)

// 	// Test exponential backoff calculation
// 	testCases := []struct {
// 		attempt  int
// 		expected time.Duration
// 	}{
// 		{1, 2 * time.Second},   // 1s * 2^1 = 2s
// 		{2, 4 * time.Second},   // 1s * 2^2 = 4s
// 		{3, 8 * time.Second},   // 1s * 2^3 = 8s
// 		{4, 16 * time.Second},  // 1s * 2^4 = 16s
// 		{5, 30 * time.Second},  // Capped at 30s
// 		{6, 30 * time.Second},  // Capped at 30s
// 		{10, 30 * time.Second}, // Capped at 30s
// 	}

// 	for _, tc := range testCases {
// 		result := sysAp.calculateBackoffDuration(tc.attempt)
// 		if result != tc.expected {
// 			t.Errorf("For attempt %d, expected %v, got %v", tc.attempt, tc.expected, result)
// 		}
// 	}
// }

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
