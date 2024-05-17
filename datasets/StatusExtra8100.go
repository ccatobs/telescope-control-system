package datasets

type StatusExtra8100 struct {
	AzimuthPositionCommand                      float64 `json:"Azimuth position command"`     // Unit: [deg]
	AzimuthCurrentAcceleration                  float64 `json:"Azimuth current acceleration"` // Unit: [deg/s^2]
	AzimuthCombinedAutotrackIntegrationFactor   float64 `json:"Azimuth Combined Autotrack Integration Factor"`
	AzimuthTorqueBiasMode                       uint8   `json:"Azimuth torque bias mode"`
	AzimuthOvercurrentMotor1                    uint8   `json:"Azimuth overcurrent motor 1"`
	AzimuthOvercurrentMotor2                    uint8   `json:"Azimuth overcurrent motor 2"`
	AzimuthOvercurrentMotor3                    uint8   `json:"Azimuth overcurrent motor 3"`
	AzimuthOvercurrentMotor4                    uint8   `json:"Azimuth overcurrent motor 4"`
	AzimuthOscillationStatus                    uint8   `json:"Azimuth oscillation status"`
	AzimuthOscillationWarning                   uint8   `json:"Azimuth oscillation warning"`
	AzimuthOscillationFailue                    uint8   `json:"Azimuth oscillation failue"`
	AzimuthProfilerActive                       bool    `json:"Azimuth profiler active"`
	AzimuthProfilerRunning                      bool    `json:"Azimuth profiler running"`
	AzimuthInterlockAlarm                       uint8   `json:"Azimuth interlock alarm"`
	AzimuthTorque                               float64 `json:"Azimuth torque"`                 // Unit: [%]
	ElevationPositionCommand                    float64 `json:"Elevation position command"`     // Unit: [deg]
	ElevationCurrentAcceleration                float64 `json:"Elevation current acceleration"` // Unit: [deg/s^2]
	ElevationCombinedAutotrackIntegrationFactor float64 `json:"Elevation Combined Autotrack Integration Factor"`
	ElevationTorqueBiasMode                     uint8   `json:"Elevation torque bias mode"`
	ElevationOvercurrentMotor1                  uint8   `json:"Elevation overcurrent motor 1"`
	ElevationOvercurrentMotor2                  uint8   `json:"Elevation overcurrent motor 2"`
	ElevationOvercurrentMotor3                  uint8   `json:"Elevation overcurrent motor 3"`
	ElevationOvercurrentMotor4                  uint8   `json:"Elevation overcurrent motor 4"`
	ElevationOscillationStatus                  uint8   `json:"Elevation oscillation status"`
	ElevationOscillationWarning                 uint8   `json:"Elevation oscillation warning"`
	ElevationOscillationFailue                  uint8   `json:"Elevation oscillation failue"`
	ElevationProfilerActive                     bool    `json:"Elevation profiler active"`
	ElevationProfilerRunning                    bool    `json:"Elevation profiler running"`
	ElevationInterlockAlarm                     uint8   `json:"Elevation interlock alarm"`
	ElevationTorque                             float64 `json:"Elevation torque"` // Unit: [%]
	ElevationStowPosition                       bool    `json:"Elevation stow position"`
	ElevationStowStatus                         uint8   `json:"Elevation stow status"`
	ElevationTrue                               float64 `json:"Elevation true"` // Unit: [deg]
	Speed                                       uint8   `json:"Speed"`
	GreenMode                                   bool    `json:"Green Mode"`
	TrackingStatus                              uint8   `json:"Tracking Status"`
	AutotrackStatus                             uint8   `json:"Autotrack Status"`
	SplineStatus                                uint8   `json:"Spline Status"`
	CommandingQuality                           uint8   `json:"Commanding Quality"`
	GeneralInterlockAlarm                       uint8   `json:"General interlock alarm"`
	QtyOfUsedProgramTrackStackPositions         int32   `json:"Qty of used program track stack positions"`
	QtyOfFreeProgramTrackStackPositions         uint16  `json:"Qty of free program track stack positions"`
	TimeToPosition                              float64 `json:"TimeToPosition"`
	TimeStampCommand                            float64 `json:"Time Stamp Command"`          // Unit: [MJD1950]
	AzimuthNormalizedPosition                   float64 `json:"Azimuth normalized position"` // Unit: [deg]
	AzimuthCanBusFailure                        uint8   `json:"Azimuth can bus failure"`
	ElevationCanBusFailure                      uint8   `json:"Elevation can bus failure"`
}
