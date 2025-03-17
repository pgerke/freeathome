package models

// VirtualDevice represents a virtual device with a type and properties.
type VirtualDevice struct {
	// Type of the virtual device.
	Type VirtualDeviceType `json:"type"`

	// Properties of the virtual device.
	Properties VirtualDeviceProperties `json:"properties"`
}

// VirtualDeviceProperties represents the properties of a virtual device.
type VirtualDeviceProperties struct {
	// TTL represents the time-to-live of the virtual device.
	TTL *string `json:"ttl,omitempty"`

	// DisplayName represents the display name of the virtual device.
	DisplayName *string `json:"displayName,omitempty"`

	// Flavor represents the flavor of the virtual device.
	Flavor *string `json:"flavor,omitempty"`

	// Capabilities represents the capabilities of the virtual device.
	Capabilities *[]uint `json:"capabilities,omitempty"`
}

// VirtualDeviceType represents the type of a virtual device.
type VirtualDeviceType int

// VirtualDeviceType constants.
const (
	Unknown VirtualDeviceType = iota
	BinarySensor
	BlindActuator
	SwitchingActuator
	CeilingFanActuator
	RTC
	DimActuator
	EVCharging
	WindowSensor
	SimpleDoorlock
	ShutterActuator
	WeatherStation
	WeatherTemperatureSensor
	WeatherWindSensor
	WeatherBrightnessSensor
	WeatherRainSensor
	WindowActuator
	CODetector
	FireDetector
	KNXSwitchSensor
	MediaPlayer
	EnergyBattery
	EnergyInverter
	EnergyMeter
	EnergyInverterBattery
	EnergyInverterMeter
	EnergyInverterMeterBattery
	EnergyMeterBattery
	AirQualityCO2
	AirQualityCO
	AirQualityFull
	AirQualityHumidity
	AirQualityNO2
	AirQualityO3
	AirQualityPM10
	AirQualityPM25
	AirQualityPressure
	AirQualityTemperature
	AirQualityVOC
	EnergyMeterV2
	HomeApplianceLaundry
	HVAC
	SplitUnit
)
