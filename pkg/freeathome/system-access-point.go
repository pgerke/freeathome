package freeathome

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"

	"github.com/pgerke/freeathome/pkg/models"
)

type connection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteControl(messageType int, data []byte, deadline time.Time) error
}

// SystemAccessPoint represents a system access point that can be used to communicate with a free@home system.
type SystemAccessPoint struct {
	UUID string
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
	// waitGroup is used to synchronize the web socket connection and message handling.
	waitGroup sync.WaitGroup
	// webSocketMessageChannel is the channel that is used to send messages received from the web socket connection.
	webSocketMessageChannel chan []byte
	// messageReceivedChannel is the channel that is used to signal that a message has been received.
	messageReceivedChannel chan struct{}
	// datapointRegex is the regular expression that is used to match datapoint keys.
	datapointRegex *regexp.Regexp
	// onMessageHandled is a callback function that is called when a message is handled.
	onMessageHandled func()
	// onError is a callback function that is called when an error occurs.
	onError func(error)
}

// NewSystemAccessPoint creates a new SystemAccessPoint with the specified host name, user name, password, TLS enabled flag, verbose errors flag, and logger.
func NewSystemAccessPoint(hostName string, userName string, password string, tlsEnabled bool, verboseErrors bool, logger models.Logger) *SystemAccessPoint {
	if logger == nil {
		logger = NewDefaultLogger(nil)
		logger.Warn("No logger provided for SystemAccessPoint. Using default logger.")
	}

	return &SystemAccessPoint{
		UUID:                    models.EmptyUUID,
		hostName:                hostName,
		logger:                  logger,
		tlsEnabled:              tlsEnabled,
		verboseErrors:           verboseErrors,
		client:                  resty.New().SetBasicAuth(userName, password),
		waitGroup:               sync.WaitGroup{},
		webSocketMessageChannel: nil,
		messageReceivedChannel:  nil,
		datapointRegex:          regexp.MustCompile(models.DatapointPattern),
	}
}

// emitError is a helper function to emit errors using the onError callback.
func (sysAp *SystemAccessPoint) emitError(err error) {
	if sysAp.onError != nil {
		sysAp.onError(err)
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
func (sysAp *SystemAccessPoint) ConnectWebSocket(ctx context.Context, keepaliveInterval time.Duration) {
	// TODO: Implement exponential backoff for reconnection attempts
	// backoff := time.Second
	// TODO: Implement a maximum duration for reconnection attempts
	// TODO: Implement a maximum number of reconnection attempts

	// Wait for all processes to finish before returning
	defer sysAp.waitGroup.Wait()

	// Start the connection loop
	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, stop the connection attempts
			sysAp.logger.Log("context cancelled, stopping web socket connection attempts")
			return
		default:
			// Attempt to establish a web socket connection
			sysAp.webSocketConnectionLoop(ctx, keepaliveInterval)
		}
	}
}

// CreateVirtualDevice creates a new virtual device on the System Access Point (SysAP) with the specified serial number.
// It sends a PUT request containing the provided VirtualDevice data to the SysAP API.
// On success, it returns the response containing details of the created virtual device.
// If an error occurs during the request or response parsing, it logs the error, emits an error event, and returns the error.
//
// Parameters:
//   - serial: The serial number of the virtual device to be created.
//   - virtualDevice: A pointer to the VirtualDevice struct containing the device configuration.
//
// Returns:
//   - *models.VirtualDeviceResponse: Pointer to the response struct with details of the created virtual device.
//   - error: An error object if the operation fails, otherwise nil.
func (sysAp *SystemAccessPoint) CreateVirtualDevice(serial string, virtualDevice *models.VirtualDevice) (*models.VirtualDeviceResponse, error) {
	resp, err := sysAp.client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial}).
		SetBody(virtualDevice).
		Put(sysAp.GetUrl("virtualdevice/{uuid}/{serial}"))

	return deserializeRestResponse[models.VirtualDeviceResponse](sysAp, resp, err, "failed to create virtual device")
}

// webSocketConnectionLoop establishes a web socket connection and starts the message loop.
func (sysAp *SystemAccessPoint) webSocketConnectionLoop(ctx context.Context, keepaliveInterval time.Duration) {
	// Add a wait group to ensure all processes are finished before returning
	sysAp.waitGroup.Add(1)
	defer sysAp.waitGroup.Done()

	// Create a new web socket connection
	basicAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", sysAp.client.UserInfo.Username, sysAp.client.UserInfo.Password)))
	conn, _, err := websocket.DefaultDialer.Dial(sysAp.getWebSocketUrl(), http.Header{
		"Authorization": []string{fmt.Sprintf("Basic %s", basicAuth)},
	})

	// Check for errors
	if err != nil {
		sysAp.logger.Error("failed to connect to web socket", "error", err)
		sysAp.emitError(err)
		// time.Sleep(backoff)
		return
	}

	// Create connection channels
	sysAp.messageReceivedChannel = make(chan struct{}, 1)
	sysAp.webSocketMessageChannel = make(chan []byte, 10)
	defer func() {
		close(sysAp.messageReceivedChannel)
		sysAp.messageReceivedChannel = nil
		close(sysAp.webSocketMessageChannel)
		sysAp.webSocketMessageChannel = nil
	}()

	// Start keepalive and message handler goroutines
	go sysAp.webSocketKeepaliveLoop(conn, keepaliveInterval)
	go sysAp.webSocketMessageHandler()

	// Start the message loop
	sysAp.logger.Log("web socket connected successfully, starting message loop")
	err = sysAp.webSocketMessageLoop(ctx, conn)

	// Check for errors
	if err != nil {
		sysAp.logger.Error("web socket message loop failed", "error", err)
		sysAp.emitError(err)
	}

	// Close the web socket connection
	err = conn.Close()
	sysAp.logger.Debug("web socket connection closed", "error", err)
}

// webSocketMessageLoop starts a loop to read messages from the web socket connection.
func (sysAp *SystemAccessPoint) webSocketMessageLoop(ctx context.Context, conn connection) error {
	// Verify that the connection channels are not nil
	if sysAp.webSocketMessageChannel == nil || sysAp.messageReceivedChannel == nil {
		errorMessage := "a connection channel is nil, cannot start message loop"
		sysAp.logger.Error(errorMessage)
		return errors.New(errorMessage)
	}

	// Start a loop to read messages from the web socket
	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, stop the message loop
			sysAp.logger.Log("context cancelled, stopping message loop")
			return nil
		default:
			// Read messages from the web socket
			messageType, message, err := conn.ReadMessage()

			// Check for errors
			if err != nil {
				sysAp.emitError(err)
				return err
			}

			// Signal that a message has been received
			sysAp.messageReceivedChannel <- struct{}{}

			// Check if the message type is text
			if messageType != websocket.TextMessage {
				sysAp.logger.Warn("received non-text message from web socket", "type", messageType, "message", string(message))
				continue
			}

			// Pipe the message to the message handler
			sysAp.logger.Debug("received text message from web socket")
			sysAp.webSocketMessageChannel <- message
		}
	}
}

// processWebSocketMessage processes a message received from the web socket connection.
func (sysAp *SystemAccessPoint) webSocketMessageHandler() {
	// Add a wait group to ensure all processes are finished before returning
	sysAp.waitGroup.Add(1)
	defer sysAp.waitGroup.Done()

	// Verify that the webSocketMessageChannel is not nil
	if sysAp.webSocketMessageChannel == nil {
		sysAp.logger.Error("webSocketMessageChannel is nil, cannot start message handler")
		return
	}

	// Start a loop to handle messages from the web socket
	for message := range sysAp.webSocketMessageChannel {
		sysAp.processMessage(message)
	}

	// If the channel is closed, exit the loop
	sysAp.logger.Log("webSocketMessageChannel closed, stopping message handler")
}

func (sysAp *SystemAccessPoint) webSocketKeepaliveLoop(conn connection, interval time.Duration) {
	// Add a wait group to ensure all processes are finished before returning
	sysAp.waitGroup.Add(1)
	defer sysAp.waitGroup.Done()

	// Verify that the messageReceivedChannel is not nil
	if sysAp.messageReceivedChannel == nil {
		sysAp.logger.Error("messageReceivedChannel is nil, cannot start keepalive loop")
		return
	}

	// Create a ticker for the keepalive interval
	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case _, ok := <-sysAp.messageReceivedChannel:
			if ok {
				// Reset the timer when a message is received
				sysAp.logger.Debug("message received, resetting keepalive timer")
				timer.Reset(interval)
			} else {
				// If the channel is closed, exit the loop
				sysAp.logger.Log("messageReceivedChannel closed, stopping keepalive")
				return
			}
		case <-timer.C:
			// Send a ping message to the server
			sysAp.logger.Log("keepalive timer expired, sending ping message...")
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(3*time.Second))
			if err != nil {
				sysAp.logger.Error("failed to send ping message", "error", err)
				sysAp.emitError(err)
				return
			}
		}
	}
}

func (sysAp *SystemAccessPoint) processMessage(message []byte) {
	defer func() {
		// Call the onMessageHandled callback if it is set
		if sysAp.onMessageHandled != nil {
			sysAp.onMessageHandled()
		}
	}()

	// Unmarshal the message into a WebSocketMessage struct
	var msg models.WebSocketMessage
	err := json.Unmarshal(message, &msg)

	if err != nil {
		sysAp.logger.Error("failed to unmarshal message", "error", err)
		sysAp.emitError(err)
		return
	}

	// Check if the message is empty
	if len(msg[models.EmptyUUID].Datapoints) == 0 {
		sysAp.logger.Warn("web socket message has no datapoints")
		return
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

// GetConfiguration retrieves the configuration from the system access point.
// It sends a GET request to the "configuration" endpoint and unmarshals the response
// into a models.Configuration object.
//
// Returns a pointer to the Configuration object and an error if any occurred during the request.
//
// Possible errors include network issues, non-2xx HTTP responses, or unmarshalling errors.
func (sysAp *SystemAccessPoint) GetConfiguration() (*models.Configuration, error) {
	resp, err := sysAp.client.R().Get(sysAp.GetUrl("configuration"))

	return deserializeRestResponse[models.Configuration](sysAp, resp, err, "failed to get configuration")
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

	return deserializeRestResponse[models.DeviceList](sysAp, resp, err, "failed to get device list")
}

// GetDevice retrieves a device with the specified serial number from the system access point.
// It sends a GET request to the appropriate endpoint and parses the response into a DeviceResponse model.
// Returns a pointer to the DeviceResponse and an error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) GetDevice(serial string) (*models.DeviceResponse, error) {
	resp, err := sysAp.client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial}).
		Get(sysAp.GetUrl("device/{uuid}/{serial}"))

	return deserializeRestResponse[models.DeviceResponse](sysAp, resp, err, "failed to get device")
}

// GetDatapoint retrieves a datapoint from the System Access Point using the provided serial number, channel, and datapoint identifiers.
// It sends a GET request to the corresponding endpoint and parses the response into a models.Datapoint object.
// If the request fails or the response cannot be parsed, an error is returned and logged.
//
// Parameters:
//
//	serial   - The serial number of the device.
//	channel  - The channel identifier.
//	datapoint - The datapoint identifier.
//
// Returns:
//
//	*models.Datapoint - The retrieved datapoint object.
//	error             - An error if the request or parsing fails.
func (sysAp *SystemAccessPoint) GetDatapoint(serial string, channel string, datapoint string) (*models.GetDataPointResponse, error) {
	resp, err := sysAp.client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial, "channel": channel, "datapoint": datapoint}).
		Get(sysAp.GetUrl("datapoint/{uuid}/{serial}.{channel}.{datapoint}"))

	return deserializeRestResponse[models.GetDataPointResponse](sysAp, resp, err, "failed to get datapoint")
}

// SetDatapoint sets the value of a specified datapoint for a given device channel.
// It sends a PUT request to the System Access Point (SysAP) to update the datapoint value.
//
// Parameters:
//
//	serial    - The serial number of the target device.
//	channel   - The channel identifier of the device.
//	datapoint - The datapoint identifier to be set.
//	value     - The value to set for the datapoint.
//
// Returns:
//
//	*models.SetDataPointResponse - The response from the SysAP after setting the datapoint.
//	error                        - An error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) SetDatapoint(serial string, channel string, datapoint string, value string) (*models.SetDataPointResponse, error) {
	resp, err := sysAp.client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial, "channel": channel, "datapoint": datapoint}).
		SetBody(value).
		Put(sysAp.GetUrl("datapoint/{uuid}/{serial}.{channel}.{datapoint}"))

	return deserializeRestResponse[models.SetDataPointResponse](sysAp, resp, err, "failed to set datapoint")
}

// TriggerProxyDevice sends a request to trigger an action on a proxy device identified by its class and serial number.
// It constructs the request URL using the SystemAccessPoint's UUID, the device class, serial, and the specified action.
// The method returns the parsed DeviceResponse on success, or an error if the request fails or the response cannot be parsed.
//
// Parameters:
//   - class:  The class of the proxy device.
//   - serial: The serial number of the proxy device.
//   - action: The action to trigger on the proxy device.
//
// Returns:
//   - *models.DeviceResponse: The response from the device if the action is successful.
//   - error: An error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) TriggerProxyDevice(class string, serial string, action string) (*models.DeviceResponse, error) {
	resp, err := sysAp.client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "class": class, "serial": serial, "action": action}).
		Get(sysAp.GetUrl("proxydevice/{uuid}/{class}/{serial}/action/{action}"))

	return deserializeRestResponse[models.DeviceResponse](sysAp, resp, err, "failed to trigger proxy device")
}

// SetProxyDeviceValue sets the value of a proxy device identified by its class and serial number.
// It sends a PUT request to the system access point's API and returns the device response.
//
// Parameters:
//   - class:  The device class identifier.
//   - serial: The serial number of the device.
//   - value:  The value to set for the device.
//
// Returns:
//   - *models.DeviceResponse: The response from the device if the operation is successful.
//   - error: An error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) SetProxyDeviceValue(class string, serial string, value string) (*models.DeviceResponse, error) {
	resp, err := sysAp.client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "class": class, "serial": serial, "value": value}).
		Put(sysAp.GetUrl("proxydevice/{uuid}/{class}/{serial}/value/{value}"))

	return deserializeRestResponse[models.DeviceResponse](sysAp, resp, err, "failed to set proxy device value")
}

func deserializeRestResponse[T any](sysAp *SystemAccessPoint, resp *resty.Response, err error, errorMessage string) (*T, error) {
	// Check for errors
	if err != nil {
		sysAp.logger.Error(errorMessage, "error", err)
		sysAp.emitError(err)
		return nil, err
	}

	if resp.IsError() {
		sysAp.logger.Error(errorMessage, "status", resp.Status(), "body", resp.String())
		return nil, fmt.Errorf("%s: %s", errorMessage, resp.String())
	}

	var object T
	if err := json.Unmarshal(resp.Body(), &object); err != nil {
		sysAp.logger.Error("failed to parse response body", "error", err)
		sysAp.emitError(err)
		return nil, err
	}

	return &object, nil
}
