package freeathome

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestSystemAccessPoint_GetDeviceList tests the GetDeviceList method of SystemAccessPoint.
func TestSystemAccessPoint_GetDeviceList(t *testing.T) {
	sysAp, buf := setup(t, true)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       loadTestResponseBody(t, "devicelist.json"),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.client.SetTransport(roundtripper)

	result, err := sysAp.GetDeviceList()

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
		t.Errorf("Expected GET request, got %s", response.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/devicelist"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected devices from one system access point, got %d", len(*result))
	}
	if len((*result)["00000000-0000-0000-0000-000000000000"]) != 76 {
		t.Errorf("Expected 76 devices, got %d", len((*result)["00000000-0000-0000-0000-000000000000"]))
	}
}

// TestSystemAccessPoint_GetDeviceListCallError tests the GetDeviceList method of SystemAccessPoint
func TestSystemAccessPoint_GetDeviceListCallError(t *testing.T) {
	sysAp, buf := setup(t, true)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.client.SetTransport(roundtripper)

	result, err := sysAp.GetDeviceList()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if strings.Contains(logOutput, "msg=failed to get device list") || strings.Contains(logOutput, "error=\"Get \"https://localhost/fhapi/v1/api/rest/devicelist\": Test Error") {
		t.Errorf("Unexpected log output', got: %s", logOutput)
	}
	// Check if result is nil and error is not nil
	if result != nil {
		t.Error("Expected nil result")
	}
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Check if the error message is correct
	expected := "Get \"https://localhost/fhapi/v1/api/rest/devicelist\": Test Error"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%v'", expected, err)

	}
}

// TestSystemAccessPoint_GetDeviceListErrorResponse tests the GetDeviceList method of SystemAccessPoint
func TestSystemAccessPoint_GetDeviceListErrorResponse(t *testing.T) {
	sysAp, buf := setup(t, true)
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

	result, err := sysAp.GetDeviceList()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get device list\"") ||
		!strings.Contains(logOutput, "level=ERROR") ||
		!strings.Contains(logOutput, "status=\"Internal Server Error\"") ||
		!strings.Contains(logOutput, "body=\"Internal Server Error\"") {
		t.Errorf("Unexpected log output', got: %s", logOutput)
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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/devicelist"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to get device list: Internal Server Error"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%v'", expected, err)
	}
}

// TestSystemAccessPoint_GetDeviceListUnmarshalError tests the GetDeviceList method of SystemAccessPoint
func TestSystemAccessPoint_GetDeviceListUnmarshalError(t *testing.T) {
	sysAp, buf := setup(t, true)
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

	result, err := sysAp.GetDeviceList()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to parse device list\"") ||
		!strings.Contains(logOutput, "level=ERROR") {
		t.Errorf("Unexpected log output', got: %s", logOutput)
	}

	// Check if result is nil and error is not nil
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Error("Expected nil result")
	}

	// Check if the error message is correct
	expected := "json: cannot unmarshal object into Go value of type string"
	if err.Error() != expected {
		t.Errorf("Expected error '%s', got '%v'", expected, err)
	}
}
