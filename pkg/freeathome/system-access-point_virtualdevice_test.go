package freeathome

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pgerke/freeathome/pkg/models"
)

func newVirtualDevice(t *testing.T) *models.VirtualDevice {
	t.Helper()
	properties := &models.VirtualDeviceProperties{
		TTL:          nil,
		DisplayName:  nil,
		Flavor:       nil,
		Capabilities: &[]uint{1, 2, 3},
	}
	return &models.VirtualDevice{
		Type:       models.BinarySensor,
		Properties: *properties,
	}
}

func TestSystemAccessPointCreateVirtualDevice(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       loadTestResponseBody(t, "virtualdevice.json"),
		Header:     make(http.Header),
	}

	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.client.SetTransport(roundtripper)

	device := newVirtualDevice(t)
	result, err := sysAp.CreateVirtualDevice("6000D2CB27B2", device)

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
	if roundtripper.Request.Method != http.MethodPut {
		t.Errorf("Expected PUT request, got %s", roundtripper.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/virtualdevice/00000000-0000-0000-0000-000000000000/6000D2CB27B2"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if result == nil || *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the response to contain one virtual device response, got %d", len(*result))
	}
	if len((*result)[models.EmptyUUID].Devices) != 1 {
		t.Errorf("Expected 1 created virtual device, got %d", len((*result)[models.EmptyUUID].Devices))
	}
	if (*result)[models.EmptyUUID].Devices["abcd12345"].Serial != "6000D2CB27B2" {
		t.Errorf("Expected created virtual device serial to be '6000D2CB27B2', got '%s'", (*result)[models.EmptyUUID].Devices["abcd12345"].Serial)
	}
}

func TestSystemAccessPointCreateVirtualDeviceCallError(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.client.SetTransport(roundtripper)

	device := newVirtualDevice(t)
	result, err := sysAp.CreateVirtualDevice("6000D2CB27B2", device)

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to create virtual device\"") ||
		!strings.Contains(logOutput, "error=\"Put \\\"https://localhost/fhapi/v1/api/rest/virtualdevice/00000000-0000-0000-0000-000000000000/6000D2CB27B2\\\": Test Error") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if err == nil {
		t.Fatal(expectedErrorGotNil)
	}

	// Check if the error message is correct
	expected := "Put \"https://localhost/fhapi/v1/api/rest/virtualdevice/00000000-0000-0000-0000-000000000000/6000D2CB27B2\": Test Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointCreateVirtualDeviceErrorResponse(t *testing.T) {
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

	device := newVirtualDevice(t)
	result, err := sysAp.CreateVirtualDevice("6000D2CB27B2", device)

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to create virtual device\"") ||
		!strings.Contains(logOutput, "level=ERROR") ||
		!strings.Contains(logOutput, "status=\"Internal Server Error\"") ||
		!strings.Contains(logOutput, "body=\"Internal Server Error\"") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if result != nil {
		t.Error(expectedNil)
	}
	if err == nil {
		t.Error(expectedErrorGotNil)
	}

	// Check if the request method and URL are correct
	if roundtripper.Request.Method != http.MethodPut {
		t.Errorf("Expected PUT request, got %s", roundtripper.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/virtualdevice/00000000-0000-0000-0000-000000000000/6000D2CB27B2"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to create virtual device: Internal Server Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointCreateVirtualDeviceUnmarshalError(t *testing.T) {
	sysAp, buf, _ := setup(t, true)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"00000000-0000-0000-0000-000000000000":{"devices":{"abcd12345":{"serial": 123}}}}`)),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.client.SetTransport(roundtripper)

	device := newVirtualDevice(t)
	result, err := sysAp.CreateVirtualDevice("6000D2CB27B2", device)

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to parse response body\"") ||
		!strings.Contains(logOutput, "level=ERROR") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if result != nil {
		t.Error(expectedNil)
	}
	if err == nil {
		t.Fatal(expectedErrorGotNil)
	}

	// Check if the error message is correct
	expected := "json: cannot unmarshal number into Go struct field CreatedVirtualDevice.devices.serial of type string"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}
