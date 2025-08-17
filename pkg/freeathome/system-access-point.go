package freeathome

import (
	"context"
	"crypto/tls"
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

// Config represents the configuration for a SystemAccessPoint
type Config struct {
	// Hostname is the hostname or IP address of the system access point
	Hostname string
	// Username is the username for authentication
	Username string
	// Password is the password for authentication
	Password string
	// TLSEnabled indicates whether TLS is enabled for communication
	TLSEnabled bool
	// SkipTLSVerify indicates whether TLS certificate verification should be skipped
	SkipTLSVerify bool
	// VerboseErrors indicates whether verbose errors should be logged
	VerboseErrors bool
	// Logger is the logger to use for logging messages
	Logger models.Logger
	// Client is the REST client to use (optional, will create default if nil)
	Client *resty.Client
}

// NewConfig creates a new Config with default values
func NewConfig(hostname, username, password string) *Config {
	return &Config{
		Hostname:      hostname,
		Username:      username,
		Password:      password,
		TLSEnabled:    true,
		SkipTLSVerify: false,
		VerboseErrors: false,
		Logger:        nil,
		Client:        nil,
	}
}

type connection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteControl(messageType int, data []byte, deadline time.Time) error
}

// SystemAccessPoint represents a system access point that can be used to communicate with a free@home system.
type SystemAccessPoint struct {
	UUID string
	// config contains the configuration for the system access point
	config *Config
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
	// reconnectionAttempts tracks the number of failed reconnection attempts
	reconnectionAttempts int
	// maxReconnectionAttempts is the maximum number of reconnection attempts before giving up
	maxReconnectionAttempts int
	// reconnectionMutex protects access to reconnectionAttempts
	reconnectionMutex sync.Mutex
	// exponentialBackoffEnabled controls whether exponential backoff is used between reconnection attempts
	exponentialBackoffEnabled bool
}

// NewSystemAccessPoint creates a new SystemAccessPoint with the specified configuration.
func NewSystemAccessPoint(config *Config) (*SystemAccessPoint, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Set default logger if not provided
	if config.Logger == nil {
		config.Logger = NewDefaultLogger(nil)
		config.Logger.Warn("No logger provided for SystemAccessPoint. Using default logger.")
	}

	// Create REST client with basic auth
	if config.Client == nil {
		config.Client = resty.New()
	}
	config.Client.SetBasicAuth(config.Username, config.Password)

	// Configure TLS settings if TLS is enabled
	if config.TLSEnabled && config.SkipTLSVerify {
		config.Logger.Warn("TLS is enabled but certificate verification is disabled, this is not recommended!")
		config.Client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return &SystemAccessPoint{
		UUID:                      models.EmptyUUID,
		config:                    config,
		waitGroup:                 sync.WaitGroup{},
		webSocketMessageChannel:   nil,
		messageReceivedChannel:    nil,
		datapointRegex:            regexp.MustCompile(models.DatapointPattern),
		reconnectionAttempts:      0,
		maxReconnectionAttempts:   3,
		reconnectionMutex:         sync.Mutex{},
		exponentialBackoffEnabled: true,
	}, nil
}

// MustNewSystemAccessPoint creates a new SystemAccessPoint with the specified configuration.
// It panics if an error occurs.
func MustNewSystemAccessPoint(config *Config) *SystemAccessPoint {
	sysap, err := NewSystemAccessPoint(config)
	// The error can only occur if the config is nil, which is considered a programming error.
	// If you are not sure if the config is nil, use NewSystemAccessPoint instead.
	if err != nil {
		panic(err)
	}
	return sysap
}

// NewSystemAccessPointWithDefaults creates a new SystemAccessPoint with minimal configuration
func NewSystemAccessPointWithDefaults(hostname, username, password string) *SystemAccessPoint {
	config := NewConfig(hostname, username, password)
	return MustNewSystemAccessPoint(config)
}

// emitError is a helper function to emit errors using the onError callback.
func (sysAp *SystemAccessPoint) emitError(err error) {
	if sysAp.onError != nil {
		sysAp.onError(err)
	}
}

// HostName returns the host name of the system access point.
func (sysAp *SystemAccessPoint) GetHostName() string {
	return sysAp.config.Hostname
}

// TlsEnabled returns whether TLS is enabled for communication with the system access point.
func (sysAp *SystemAccessPoint) GetTlsEnabled() bool {
	return sysAp.config.TLSEnabled
}

// SkipTLSVerify returns whether TLS certificate verification should be skipped.
func (sysAp *SystemAccessPoint) GetSkipTLSVerify() bool {
	return sysAp.config.SkipTLSVerify
}

// VerboseErrors returns whether verbose errors should be logged.
func (sysAp *SystemAccessPoint) GetVerboseErrors() bool {
	return sysAp.config.VerboseErrors
}

// SetMaxReconnectionAttempts sets the maximum number of reconnection attempts.
func (sysAp *SystemAccessPoint) SetMaxReconnectionAttempts(maxAttempts int) {
	sysAp.reconnectionMutex.Lock()
	defer sysAp.reconnectionMutex.Unlock()
	sysAp.maxReconnectionAttempts = maxAttempts
}

// GetMaxReconnectionAttempts returns the maximum number of reconnection attempts.
func (sysAp *SystemAccessPoint) GetMaxReconnectionAttempts() int {
	sysAp.reconnectionMutex.Lock()
	defer sysAp.reconnectionMutex.Unlock()
	return sysAp.maxReconnectionAttempts
}

// GetReconnectionAttempts returns the current number of reconnection attempts.
func (sysAp *SystemAccessPoint) GetReconnectionAttempts() int {
	sysAp.reconnectionMutex.Lock()
	defer sysAp.reconnectionMutex.Unlock()
	return sysAp.reconnectionAttempts
}

// SetExponentialBackoffEnabled sets whether exponential backoff is enabled for reconnection attempts.
func (sysAp *SystemAccessPoint) SetExponentialBackoffEnabled(enabled bool) {
	sysAp.reconnectionMutex.Lock()
	defer sysAp.reconnectionMutex.Unlock()
	sysAp.exponentialBackoffEnabled = enabled
}

// GetExponentialBackoffEnabled returns whether exponential backoff is enabled for reconnection attempts.
func (sysAp *SystemAccessPoint) GetExponentialBackoffEnabled() bool {
	sysAp.reconnectionMutex.Lock()
	defer sysAp.reconnectionMutex.Unlock()
	return sysAp.exponentialBackoffEnabled
}

// calculateBackoffDuration calculates the exponential backoff duration for a given attempt number.
// The backoff follows the formula: baseDelay * (2^attempt) with a maximum cap.
func (sysAp *SystemAccessPoint) calculateBackoffDuration(attempt int) time.Duration {
	baseDelay := time.Second
	maxDelay := 30 * time.Second

	// Calculate exponential backoff: baseDelay * (2^attempt)
	backoffDuration := min(
		baseDelay*time.Duration(1<<attempt),
		maxDelay,
	)

	return backoffDuration
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
	if sysAp.config.TLSEnabled {
		protocol = "https"
	} else {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s/fhapi/v1/api/rest/%s", protocol, sysAp.config.Hostname, path)
}

// GetWebSocketUrl constructs a WebSocket URL string for the SystemAccessPoint.
func (sysAp *SystemAccessPoint) getWebSocketUrl() string {
	var protocol string
	if sysAp.config.TLSEnabled {
		protocol = "wss"
	} else {
		protocol = "ws"
	}
	return fmt.Sprintf("%s://%s/fhapi/v1/api/ws", protocol, sysAp.config.Hostname)
}

// ConnectWebSocket establishes a web socket connection to the system access point.
func (sysAp *SystemAccessPoint) ConnectWebSocket(ctx context.Context, keepaliveInterval time.Duration) error {
	sysAp.reconnectionMutex.Lock()
	sysAp.reconnectionAttempts = 0
	sysAp.reconnectionMutex.Unlock()

	// Start the connection loop
	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, stop the connection attempts
			sysAp.config.Logger.Log("context cancelled, stopping web socket connection attempts")
			return ctx.Err()
		default:
			// Check if we've exceeded the maximum reconnection attempts
			sysAp.reconnectionMutex.Lock()
			currentAttempts := sysAp.reconnectionAttempts
			maxAttempts := sysAp.maxReconnectionAttempts
			sysAp.reconnectionMutex.Unlock()

			if currentAttempts >= maxAttempts {
				sysAp.config.Logger.Error("maximum reconnection attempts exceeded", "attempts", currentAttempts, "max", maxAttempts)
				return errors.New("maximum reconnection attempts exceeded")
			}

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
	resp, err := sysAp.config.Client.R().
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

	// Create a custom dialer for WebSocket connection
	dialer := websocket.DefaultDialer
	if sysAp.config.TLSEnabled && sysAp.config.SkipTLSVerify {
		dialer = &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Create a new web socket connection
	basicAuth := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", sysAp.config.Client.UserInfo.Username, sysAp.config.Client.UserInfo.Password))
	conn, _, err := dialer.Dial(sysAp.getWebSocketUrl(), http.Header{
		"Authorization": []string{fmt.Sprintf("Basic %s", basicAuth)},
	})

	// Check for errors
	if err != nil {
		sysAp.reconnectionMutex.Lock()
		sysAp.reconnectionAttempts++
		currentAttempts := sysAp.reconnectionAttempts
		maxAttempts := sysAp.maxReconnectionAttempts
		backoffEnabled := sysAp.exponentialBackoffEnabled
		sysAp.reconnectionMutex.Unlock()

		// Prepare error message with backoff information
		errorAttrs := []any{"error", err, "attempt", currentAttempts, "max", maxAttempts}
		if backoffEnabled && currentAttempts < maxAttempts {
			backoffDuration := sysAp.calculateBackoffDuration(currentAttempts)
			errorAttrs = append(errorAttrs, "backoff", backoffDuration)
		}

		sysAp.config.Logger.Error("failed to connect to web socket", errorAttrs...)
		sysAp.emitError(err)

		// Apply exponential backoff if enabled
		if backoffEnabled && currentAttempts < maxAttempts {
			backoffDuration := sysAp.calculateBackoffDuration(currentAttempts)
			time.Sleep(backoffDuration)
		}

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

	// Reset reconnection attempts on successful connection
	sysAp.reconnectionMutex.Lock()
	sysAp.reconnectionAttempts = 0
	sysAp.reconnectionMutex.Unlock()

	// Start the message loop
	sysAp.config.Logger.Log("web socket connected successfully, starting message loop")
	err = sysAp.webSocketMessageLoop(ctx, conn)

	// Check for errors
	if err != nil {
		sysAp.config.Logger.Error("web socket message loop failed", "error", err)
		sysAp.emitError(err)
	}

	// Close the web socket connection
	err = conn.Close()
	sysAp.config.Logger.Debug("web socket connection closed", "error", err)
}

// webSocketMessageLoop starts a loop to read messages from the web socket connection.
func (sysAp *SystemAccessPoint) webSocketMessageLoop(ctx context.Context, conn connection) error {
	// Verify that the connection channels are not nil
	if sysAp.webSocketMessageChannel == nil || sysAp.messageReceivedChannel == nil {
		errorMessage := "a connection channel is nil, cannot start message loop"
		sysAp.config.Logger.Error(errorMessage)
		return errors.New(errorMessage)
	}

	// Start a loop to read messages from the web socket
	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, stop the message loop
			sysAp.config.Logger.Log("context cancelled, stopping message loop")
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
			select {
			case sysAp.messageReceivedChannel <- struct{}{}:
				// Message sent successfully
			case <-ctx.Done():
				// Context cancelled, exit immediately
				return ctx.Err()
			}

			// Check if the message type is text
			if messageType != websocket.TextMessage {
				sysAp.config.Logger.Warn("received non-text message from web socket", "type", messageType, "message", string(message))
				continue
			}

			// Pipe the message to the message handler
			sysAp.config.Logger.Debug("received text message from web socket")
			select {
			case sysAp.webSocketMessageChannel <- message:
				// Message sent successfully
			case <-ctx.Done():
				// Context cancelled, exit immediately
				return ctx.Err()
			}
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
		sysAp.config.Logger.Error("webSocketMessageChannel is nil, cannot start message handler")
		return
	}

	// Start a loop to handle messages from the web socket
	for message := range sysAp.webSocketMessageChannel {
		sysAp.processMessage(message)
	}

	// If the channel is closed, exit the loop
	sysAp.config.Logger.Log("webSocketMessageChannel closed, stopping message handler")
}

func (sysAp *SystemAccessPoint) webSocketKeepaliveLoop(conn connection, interval time.Duration) {
	// Add a wait group to ensure all processes are finished before returning
	sysAp.waitGroup.Add(1)
	defer sysAp.waitGroup.Done()

	// Verify that the messageReceivedChannel is not nil
	if sysAp.messageReceivedChannel == nil {
		sysAp.config.Logger.Error("messageReceivedChannel is nil, cannot start keepalive loop")
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
				sysAp.config.Logger.Debug("message received, resetting keepalive timer")
				timer.Reset(interval)
			} else {
				// If the channel is closed, exit the loop
				sysAp.config.Logger.Log("messageReceivedChannel closed, stopping keepalive")
				return
			}
		case <-timer.C:
			// Send a ping message to the server
			sysAp.config.Logger.Log("keepalive timer expired, sending ping message...")
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(3*time.Second))
			if err != nil {
				sysAp.config.Logger.Error("failed to send ping message", "error", err)
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
		sysAp.config.Logger.Error("failed to unmarshal message", "error", err)
		sysAp.emitError(err)
		return
	}

	// Check if the message is empty
	if len(msg[models.EmptyUUID].Datapoints) == 0 {
		sysAp.config.Logger.Warn("web socket message has no datapoints")
		return
	}

	// Process data point updates
	for key, datapoint := range msg[models.EmptyUUID].Datapoints {
		// Check if the key matches the expected format
		if !sysAp.datapointRegex.MatchString(key) {
			sysAp.config.Logger.Warn(`Ignored datapoint with invalid key format`, "key", key)
			continue
		}

		// Log the datapoint update
		sysAp.config.Logger.Log("data point update",
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
	resp, err := sysAp.config.Client.R().Get(sysAp.GetUrl("configuration"))

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
	resp, err := sysAp.config.Client.R().Get(sysAp.GetUrl("devicelist"))

	return deserializeRestResponse[models.DeviceList](sysAp, resp, err, "failed to get device list")
}

// GetDevice retrieves a device with the specified serial number from the system access point.
// It sends a GET request to the appropriate endpoint and parses the response into a DeviceResponse model.
// Returns a pointer to the DeviceResponse and an error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) GetDevice(serial string) (*models.DeviceResponse, error) {
	resp, err := sysAp.config.Client.R().
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
	resp, err := sysAp.config.Client.R().
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
	resp, err := sysAp.config.Client.R().
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
	resp, err := sysAp.config.Client.R().
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
	resp, err := sysAp.config.Client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "class": class, "serial": serial, "value": value}).
		Put(sysAp.GetUrl("proxydevice/{uuid}/{class}/{serial}/value/{value}"))

	return deserializeRestResponse[models.DeviceResponse](sysAp, resp, err, "failed to set proxy device value")
}

func deserializeRestResponse[T any](sysAp *SystemAccessPoint, resp *resty.Response, err error, errorMessage string) (*T, error) {
	// Check for errors
	if err != nil {
		sysAp.config.Logger.Error(errorMessage, "error", err)
		sysAp.emitError(err)
		return nil, err
	}

	if resp.IsError() {
		sysAp.config.Logger.Error(errorMessage, "status", resp.Status(), "body", resp.String())
		return nil, fmt.Errorf("%s: %s", errorMessage, resp.String())
	}

	var object T
	if err := json.Unmarshal(resp.Body(), &object); err != nil {
		sysAp.config.Logger.Error("failed to parse response body", "error", err)
		sysAp.emitError(err)
		return nil, err
	}

	return &object, nil
}
