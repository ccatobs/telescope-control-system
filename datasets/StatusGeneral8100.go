package datasets

type StatusGeneral8100 struct {
	Time                                float64 `json:"Time"`
	Year                                uint32  `json:"Year"`
	AzimuthMode                         uint8   `json:"Azimuth Mode"`
	AzimuthCommandedPosition            float64 `json:"Azimuth commanded position"`     // Unit: [deg]
	AzimuthCurrentPosition              float64 `json:"Azimuth current position"`       // Unit: [deg]
	AzimuthCurrentVelocity              float64 `json:"Azimuth current velocity"`       // Unit: [deg/s]
	AzimuthAveragePositionError         float64 `json:"Azimuth average position error"` // Unit: [deg]
	AzimuthPeakPositionError            float64 `json:"Azimuth peak position error"`    // Unit: [deg]
	AzimuthComputerDisabled             uint8   `json:"Azimuth computer disabled"`
	AzimuthAxisDisabled                 uint8   `json:"Azimuth axis disabled"`
	AzimuthAxisInStop                   bool    `json:"Azimuth axis in stop"`
	AzimuthBrakesReleased               bool    `json:"Azimuth brakes released"`
	AzimuthAxisInStowPosition           bool    `json:"Azimuth axis in stow position"`
	AzimuthStowPinsStatus               uint8   `json:"Azimuth stow pins - status"`
	AzimuthStopAtLCP                    bool    `json:"Azimuth stop at LCP"`
	AzimuthPowerOn                      bool    `json:"Azimuth power on"`
	AzimuthCCWLimit                     uint8   `json:"Azimuth CCW limit"`
	AzimuthCWLimit                      uint8   `json:"Azimuth CW limit"`
	AzimuthSummaryFault                 uint8   `json:"Azimuth summary fault"`
	ElevationMode                       uint8   `json:"Elevation Mode"`
	ElevationCommandedPosition          float64 `json:"Elevation commanded position"`     // Unit: [deg]
	ElevationCurrentPosition            float64 `json:"Elevation current position"`       // Unit: [deg]
	ElevationCurrentVelocity            float64 `json:"Elevation current velocity"`       // Unit: [deg/s]
	ElevationAveragePositionError       float64 `json:"Elevation average position error"` // Unit: [deg]
	ElevationPeakPositionError          float64 `json:"Elevation peak position error"`    // Unit: [deg]
	ElevationComputerDisabled           uint8   `json:"Elevation computer disabled"`
	ElevationAxisDisabled               uint8   `json:"Elevation axis disabled"`
	ElevationAxisInStop                 bool    `json:"Elevation axis in stop"`
	ElevationBrakesReleased             bool    `json:"Elevation brakes released"`
	ElevationAxisInStowPosition         bool    `json:"Elevation axis in stow position"`
	ElevationStowPinsStatus             uint8   `json:"Elevation stow pins - status"`
	ElevationStopAtLCP                  bool    `json:"Elevation stop at LCP"`
	ElevationPowerOn                    bool    `json:"Elevation power on"`
	ElevationCCWLimit                   uint8   `json:"Elevation CCW limit"`
	ElevationCWLimit                    uint8   `json:"Elevation CW limit"`
	ElevationSummaryFault               uint8   `json:"Elevation summary fault"`
	KeySwitchBypassEmergencyLimits      uint8   `json:"Key Switch Bypass Emergency Limits"`
	PCUOperation                        uint8   `json:"PCU Operation"`
	ATLockOn                            bool    `json:"AT Lock On"`
	Remote                              bool    `json:"Remote"`
	QtyOfFreeProgramTrackStackPositions int32   `json:"Qty of free program track stack positions"`
}
