package datasets

type CmdTwoLineTransfer struct {
	NameOfSpacecraft [32]byte `json:"Name of Spacecraft"`
	Line1            [70]byte `json:"Line 1"`
	Line2            [70]byte `json:"Line 2"`
	AOSTime          float64  `json:"AOS Time"`
	AOSAzimuth       float64  `json:"AOS Azimuth"`   // Unit: [deg]
	AOSElevation     float64  `json:"AOS Elevation"` // Unit: [deg]
	LOSTime          float64  `json:"LOS Time"`
	LOSAzimuth       float64  `json:"LOS Azimuth"`   // Unit: [deg]
	LOSElevation     float64  `json:"LOS Elevation"` // Unit: [deg]
	MAXTime          float64  `json:"MAX Time"`
	MAXAzimuth       float64  `json:"MAX Azimuth"`   // Unit: [deg]
	MAXElevation     float64  `json:"MAX Elevation"` // Unit: [deg]
}
