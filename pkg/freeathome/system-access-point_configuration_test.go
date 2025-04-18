package freeathome

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pgerke/freeathome/pkg/models"
)

// TestSystemAccessPoint_GetConfiguration tests the GetConfiguration method of SystemAccessPoint.
func TestSystemAccessPoint_GetConfiguration(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       loadTestResponseBody(t, "configuration.json"),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check for errors
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the request method and URL are correct
	if roundtripper.Request.Method != http.MethodGet {
		t.Errorf("Expected GET request, got %s", roundtripper.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/configuration"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the configuration to contain one system access point, got %d", len(*result))
	}
	if len((*result)[models.EmptyUUID].Devices) != 76 {
		t.Errorf("Expected 76 devices, got %d", len((*result)[models.EmptyUUID].Devices))
	}
}

// TestSystemAccessPoint_GetConfigurationCallError tests the GetConfiguration method of SystemAccessPoint
func TestSystemAccessPoint_GetConfigurationCallError(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if strings.Contains(logOutput, "msg=failed to get configuration") ||
		strings.Contains(logOutput, "error=\"Get \"https://localhost/fhapi/v1/api/rest/configuration\": Test Error") {
		t.Errorf("Unexpected log output, got: %s", logOutput)
	}
	// Check if result is nil and error is not nil
	if result != nil {
		t.Error("Expected nil result")
	}
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Check if the error message is correct
	expected := "Get \"https://localhost/fhapi/v1/api/rest/configuration\": Test Error"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%v'", expected, err)
	}
}

// TestSystemAccessPoint_GetConfigurationErrorResponse tests the GetConfiguration method of SystemAccessPoint
func TestSystemAccessPoint_GetConfigurationErrorResponse(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	response := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "Internal Server Error",
		Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get configuration\"") ||
		!strings.Contains(logOutput, "level=ERROR") ||
		!strings.Contains(logOutput, "status=\"Internal Server Error\"") ||
		!strings.Contains(logOutput, "body=\"Internal Server Error\"") {
		t.Errorf("Unexpected log output, got: %s", logOutput)
	}

	// Check if result is nil and error is not nil
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result")
	}

	// Check if the request method and URL are correct
	if roundtripper.Request.Method != http.MethodGet {
		t.Errorf("Expected GET request, got %s", response.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/configuration"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to get configuration: Internal Server Error"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%v'", expected, err)
	}
}

// TestSystemAccessPoint_GetConfigurationUnmarshalError tests the GetConfiguration method of SystemAccessPoint
func TestSystemAccessPoint_GetConfigurationUnmarshalError(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	response := &http.Response{
		StatusCode: http.StatusOK,
		// This is intentionally malformed JSON to trigger the unmarshal error
		Body:   io.NopCloser(strings.NewReader("{\"devices\": [{\"id\": \"device1\"}, {\"id\": \"device2\"}]}")),
		Header: make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to parse configuration\"") ||
		!strings.Contains(logOutput, "level=ERROR") {
		t.Errorf("Unexpected log output, got: %s", logOutput)
	}

	// Check if result is nil and error is not nil
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result")
	}

	// Check if the error message is correct
	expected := "json: cannot unmarshal array into Go value of type models.SysAP"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%v'", expected, err)
	}
}
