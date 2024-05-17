package datasets

type SunAvoidance struct {
	SASDeactivatedByKeySwitch           uint8   `json:"SAS deactivated by key switch"`
	SunAvoidanceInterlockEnabled        uint8   `json:"Sun avoidance Interlock enabled"`
	SunAvoidanceInterlockHasTriggered   uint8   `json:"Sun avoidance Interlock has triggered"`
	AntennaMovingDueToSunInterlock      uint8   `json:"Antenna moving due to sun Interlock"`
	SunInterlockAntennaCannotMove       uint8   `json:"Sun Interlock - antenna cannot move"`
	MinimumDistanceToSun                float64 `json:"Minimum distance to sun"`                    // Unit: [deg]
	MaximumRangeOfSunInterlock          float64 `json:"Maximum range of sun interlock"`             // Unit: [deg]
	MinimumVelocityToCrossSun           float64 `json:"Minimum velocity to cross sun"`              // Unit: [deg/s]
	PermittedTemperatureRise            uint32  `json:"Permitted temperature rise"`                 // Unit: [K]
	PermittedTimeDeviationBetweenPLCACU uint32  `json:"Permitted time deviation between PLC - ACU"` // Unit: [s]
	CoefficientTka                      float64 `json:"Coefficient tka"`
	CoefficientTkb                      float64 `json:"Coefficient tkb"`
	CoefficientTkc                      float64 `json:"Coefficient tkc"`
	CoefficientTkd                      float64 `json:"Coefficient tkd"`
	Year                                uint32  `json:"Year"`
	Month                               uint32  `json:"Month"`
	Day                                 uint32  `json:"Day"`
	Hour                                uint32  `json:"Hour"`
	Minute                              uint32  `json:"Minute"`
	Second                              uint32  `json:"Second"`
	Nanosecond                          uint32  `json:"Nanosecond"` // Unit: [ns]
}
