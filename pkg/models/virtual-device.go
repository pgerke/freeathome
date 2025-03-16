package models

type VirtualDevice struct {
	Type       VirtualDeviceType       `json:"type"`
	Properties VirtualDeviceProperties `json:"properties"`
}

type VirtualDeviceProperties struct {
	TTL *string `json:"ttl,omitempty"`

	DisplayName *string `json:"displayName,omitempty"`

	Flavor *string `json:"flavor,omitempty"`

	Capabilities *[]uint `json:"capabilities,omitempty"`
}

type VirtualDeviceType int

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
