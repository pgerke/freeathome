package freeathome

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"

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
	// webSocket is the WebSocket connection that is used to communicate with the system access point.
	// webSocket *websocket.Conn
}

// NewSystemAccessPoint creates a new SystemAccessPoint with the specified host name, user name, password, TLS enabled flag, verbose errors flag, and logger.
func NewSystemAccessPoint(hostName string, userName string, password string, tlsEnabled bool, verboseErrors bool, logger models.Logger) *SystemAccessPoint {
	if logger == nil {
		logger = NewDefaultLogger(nil)
		logger.Warn("No logger provided for SystemAccessPoint. Using default logger.")
	}

	return &SystemAccessPoint{
		hostName:      hostName,
		logger:        logger,
		tlsEnabled:    tlsEnabled,
		verboseErrors: verboseErrors,
		client:        resty.New().SetBasicAuth(userName, password),
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
