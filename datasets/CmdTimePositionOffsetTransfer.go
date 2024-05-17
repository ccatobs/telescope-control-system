package datasets

type CmdTimePositionOffsetTransfer struct {
	TimeOffset      float64 `json:"Time Offset"`      // Unit: [s]
	AzimuthOffset   float64 `json:"Azimuth Offset"`   // Unit: [deg]
	ElevationOffset float64 `json:"Elevation Offset"` // Unit: [deg]
}
