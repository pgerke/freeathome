package freeathome

import (
	"encoding/base64"

	"github.com/pgerke/freeathome/pkg/models"
)

// SystemAccessPoint represents a system access point that can be used to communicate with a free@home system.
type SystemAccessPoint struct {
	// basicAuthKey is the base64-encoded basic authentication key that is used to authenticate with the system access point.
	basicAuthKey string
	// hostName is the host name of the system access point.
	hostName string
	// logger is the logger that is used to log messages.
	logger models.Logger
	// tlsEnabled indicates whether TLS is enabled for communication with the system access point.
	tlsEnabled bool
	// verboseErrors indicates whether verbose errors should be logged.
	verboseErrors bool
}

// NewSystemAccessPoint creates a new SystemAccessPoint with the specified host name, user name, password, TLS enabled flag, verbose errors flag, and logger.
func NewSystemAccessPoint(hostName string, userName string, password string, tlsEnabled bool, verboseErrors bool, logger models.Logger) *SystemAccessPoint {
	if logger == nil {
		logger = &DefaultLogger{}
		logger.Warn("No logger provided for SystemAccessPoint. Using default logger.")
	}

	return &SystemAccessPoint{
		basicAuthKey:  base64.StdEncoding.EncodeToString([]byte(userName + ":" + password)),
		hostName:      hostName,
		logger:        logger,
		tlsEnabled:    tlsEnabled,
		verboseErrors: verboseErrors,
	}
}

// BasicAuthKey returns the base64-encoded basic authentication key that is used to authenticate with the system access point.
func (sysAp *SystemAccessPoint) GetBasicAuthKey() string {
	return sysAp.basicAuthKey
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
