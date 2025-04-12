package freeathome

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"

	"github.com/pgerke/freeathome/pkg/models"
)

// SystemAccessPoint represents a system access point that can be used to communicate with a free@home system.
type SystemAccessPoint struct {
	// hostName is the host name of the system access point.
	hostName string
	// logger is the logger that is used to log messages.
	logger models.Logger
	// tlsEnabled indicates whether TLS is enabled for communication with the system access point.
	tlsEnabled bool
	// verboseErrors indicates whether verbose errors should be logged.
	verboseErrors bool
	// client is the REST client that is used to communicate with the system access point.
	client *resty.Client
	// webSocketMessageChannel is the channel that is used to send messages received from the web socket connection.
	webSocketMessageChannel chan []byte
	// datapointRegex is the regular expression that is used to match datapoint keys.
	datapointRegex *regexp.Regexp
}

// NewSystemAccessPoint creates a new SystemAccessPoint with the specified host name, user name, password, TLS enabled flag, verbose errors flag, and logger.
func NewSystemAccessPoint(hostName string, userName string, password string, tlsEnabled bool, verboseErrors bool, logger models.Logger) *SystemAccessPoint {
	if logger == nil {
		logger = NewDefaultLogger(nil)
		logger.Warn("No logger provided for SystemAccessPoint. Using default logger.")
	}

	return &SystemAccessPoint{
		hostName:                hostName,
		logger:                  logger,
		tlsEnabled:              tlsEnabled,
		verboseErrors:           verboseErrors,
		client:                  resty.New().SetBasicAuth(userName, password),
		webSocketMessageChannel: make(chan []byte, 100),
		datapointRegex:          regexp.MustCompile(models.DatapointPattern),
	}
}

// HostName returns the host name of the system access point.
func (sysAp *SystemAccessPoint) GetHostName() string {
	return sysAp.hostName
}

// TlsEnabled returns whether TLS is enabled for communication with the system access point.
func (sysAp *SystemAccessPoint) GetTlsEnabled() bool {
	return sysAp.tlsEnabled
}

// VerboseErrors returns whether verbose errors should be logged.
func (sysAp *SystemAccessPoint) GetVerboseErrors() bool {
	return sysAp.verboseErrors
}

// GetUrl constructs a URL string for the SystemAccessPoint based on the provided path.
// It uses the appropriate protocol (http or https) depending on whether TLS is enabled.
//
// Parameters:
//   - path: The specific API endpoint path to be appended to the base URL.
//
// Returns:
//   - A formatted URL string that includes the protocol, hostname, and the provided path.
func (sysAp *SystemAccessPoint) GetUrl(path string) string {
	var protocol string
	if sysAp.tlsEnabled {
		protocol = "https"
	} else {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s/fhapi/v1/api/rest/%s", protocol, sysAp.hostName, path)
}

// GetWebSocketUrl constructs a WebSocket URL string for the SystemAccessPoint.
func (sysAp *SystemAccessPoint) getWebSocketUrl() string {
	var protocol string
	if sysAp.tlsEnabled {
		protocol = "wss"
	} else {
		protocol = "ws"
	}
	return fmt.Sprintf("%s://%s/fhapi/v1/api/ws", protocol, sysAp.hostName)
}

// ConnectWebSocket establishes a web socket connection to the system access point.
func (sysAp *SystemAccessPoint) ConnectWebSocket(ctx context.Context) error {
	// TODO: Implement exponential backoff for reconnection attempts
	// backoff := time.Second
	// TODO: Implement a maximum duration for reconnection attempts
	// TODO: Implement a maximum number of reconnection attempts
	// TODO: Send a ping message to the server every 30 seconds to avoid idle timeouts (guard timer)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", sysAp.client.UserInfo.Username, sysAp.client.UserInfo.Password)))
	// Start the message handler in a separate goroutine
	go sysAp.webSocketMessageHandler(ctx)

	for {
		select {
		case <-ctx.Done():
			// Check if the context is cancelled. In this case, don't return an error, because this is the expected way the context is terminated.
			if ctx.Err() == context.Canceled {
				sysAp.logger.Log("context cancelled, stopping web socket connection attempts")
				return nil
			}

			return ctx.Err()
		default:
			// Create a new web socket connection
			conn, _, err := websocket.DefaultDialer.Dial(sysAp.getWebSocketUrl(), http.Header{
				"Authorization": []string{fmt.Sprintf("Basic %s", basicAuth)},
			})

			// Check for errors
			if err != nil {
				sysAp.logger.Error("failed to connect to web socket", "error", err)
				// time.Sleep(backoff)
				continue
			}

			// Start the message loop
			sysAp.logger.Log("web socket connected successfully, starting message loop")
			err = sysAp.webSocketMessageLoop(ctx, conn)

			if err != nil {
				sysAp.logger.Error("web socket message loop failed", "error", err)
			}

			// Close the web socket connection
			err = conn.Close()
			if err != nil {
				sysAp.logger.Error("failed to close web socket", "error", err)
			} else {
				sysAp.logger.Debug("web socket closed successfully")
			}
		}
	}
}

// webSocketMessageLoop starts a loop to read messages from the web socket connection.
func (sysAp *SystemAccessPoint) webSocketMessageLoop(ctx context.Context, conn *websocket.Conn) error {
	// Start a loop to read messages from the web socket
	for {
		select {
		case <-ctx.Done():
			// Check if the context is cancelled. In this case, don't return an error, because this is the expected way the context is terminated.
			if ctx.Err() == context.Canceled {
				sysAp.logger.Log("context cancelled, stopping message loop")
				return nil
			}

			return ctx.Err()
		default:
			// Read messages from the web socket
			messageType, message, err := conn.ReadMessage()

			// Check for errors
			if err != nil {
				return err
			}

			// Check if the message type is text
			if messageType != websocket.TextMessage {
				sysAp.logger.Warn("received non-text message from web socket", "type", messageType)
				continue
			}

			// Pipe the message to the message handler
			sysAp.logger.Debug("Received text message from web socket")
			sysAp.webSocketMessageChannel <- message
		}
	}
}

// processWebSocketMessage processes a message received from the web socket connection.
func (sysAp *SystemAccessPoint) webSocketMessageHandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			sysAp.logger.Log("context cancelled, stopping message handler")
			return
		case message := <-sysAp.webSocketMessageChannel:
			// Unmarshal the message into a WebSocketMessage struct
			var msg models.WebSocketMessage
			err := json.Unmarshal(message, &msg)

			if err != nil {
				sysAp.logger.Error("failed to unmarshal message", "error", err)
				continue
			}

			// Check if the message is empty
			if len(msg[models.EmptyUUID].Datapoints) == 0 {
				sysAp.logger.Warn("web socket message has no datapoints")
				continue
			}

			// Process data point updates
			for key, datapoint := range msg[models.EmptyUUID].Datapoints {
				// Check if the key matches the expected format
				if !sysAp.datapointRegex.MatchString(key) {
					sysAp.logger.Warn(`Ignored datapoint with invalid key format`, "key", key)
					continue
				}

				// Log the datapoint update
				sysAp.logger.Log("data point update",
					"device", sysAp.datapointRegex.FindStringSubmatch(key)[1],
					"channel", sysAp.datapointRegex.FindStringSubmatch(key)[2],
					"datapoint", sysAp.datapointRegex.FindStringSubmatch(key)[3],
					"value", datapoint,
				)
			}
		}
	}
}

// GetConfiguration retrieves the configuration from the system access point.
// It sends a GET request to the "configuration" endpoint and unmarshals the response
// into a models.Configuration object.
//
// Returns a pointer to the Configuration object and an error if any occurred during the request.
//
// Possible errors include network issues, non-2xx HTTP responses, or unmarshalling errors.
func (sysAp *SystemAccessPoint) GetConfiguration() (*models.Configuration, error) {
	resp, err := sysAp.client.R().Get(sysAp.GetUrl("configuration"))

	// Check for errors
	if err != nil {
		sysAp.logger.Error("failed to get configuration", "error", err)
		return nil, err
	}

	if resp.IsError() {
		sysAp.logger.Error("failed to get configuration", "status", resp.Status(), "body", resp.String())
		return nil, fmt.Errorf("failed to get configuration: %s", resp.String())
	}

	var configuration models.Configuration
	if err := json.Unmarshal(resp.Body(), &configuration); err != nil {
		sysAp.logger.Error("failed to parse configuration", "error", err)
		return nil, err
	}

	return &configuration, nil
}

// GetDeviceList retrieves the list of devices from the system access point.
// It sends a GET request to the "devicelist" endpoint and unmarshals the response
// into a DeviceList model.
//
// Returns:
//   - *models.DeviceList: A pointer to the DeviceList model containing the list of devices.
//   - error: An error if the request fails or the response contains an error.
func (sysAp *SystemAccessPoint) GetDeviceList() (*models.DeviceList, error) {
	resp, err := sysAp.client.R().Get(sysAp.GetUrl("devicelist"))

	// Check for errors
	if err != nil {
		sysAp.logger.Error("failed to get device list", "error", err)
		return nil, err
	}

	if resp.IsError() {
		sysAp.logger.Error("failed to get device list", "status", resp.Status(), "body", resp.String())
		return nil, fmt.Errorf("failed to get device list: %s", resp.String())
	}

	var deviceList models.DeviceList
	if err := json.Unmarshal(resp.Body(), &deviceList); err != nil {
		sysAp.logger.Error("failed to parse device list", "error", err)
		return nil, err
	}

	return &deviceList, nil
}
