package datasets

type CmdWeatherStation struct {
	Time             float64 `json:"Time"`
	Year             uint32  `json:"Year"`
	Temperature      float64 `json:"Temperature"`      // Unit: [deg C]
	RelativeHumidity float64 `json:"RelativeHumidity"` // Unit: [%]
	AirPressure      float64 `json:"AirPressure"`      // Unit: [hPa]
}
