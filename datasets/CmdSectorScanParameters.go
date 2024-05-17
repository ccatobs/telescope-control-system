package datasets

type CmdSectorScanParameters struct {
	StartingPositionAzimuth   float64 `json:"Starting position Azimuth"`   // Unit: [deg]
	EndPositionAzimuth        float64 `json:"End position Azimuth"`        // Unit: [deg]
	LineDistanceAzimuth       float64 `json:"Line distance Azimuth"`       // Unit: [deg]
	ScanVelocityAzimuth       float64 `json:"Scan velocity Azimuth"`       // Unit: [deg/s]
	StartingPositionElevation float64 `json:"Starting position Elevation"` // Unit: [deg]
	EndPositionElevation      float64 `json:"End position Elevation"`      // Unit: [deg]
	LineDistanceElevation     float64 `json:"Line distance Elevation"`     // Unit: [deg]
	ScanVelocityElevation     float64 `json:"Scan velocity Elevation"`     // Unit: [deg/s]
	PositionWindowAzimuth     float64 `json:"Position Window Azimuth"`     // Unit: [deg]
	PositionWindowElevation   float64 `json:"Position Window Elevation"`   // Unit: [deg]
}
