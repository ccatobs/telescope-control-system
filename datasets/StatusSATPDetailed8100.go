package datasets

type StatusSATPDetailed8100 struct {
	Time                                          float64 `json:"Time"`
	Year                                          uint32  `json:"Year"`
	AzimuthMode                                   uint8   `json:"Azimuth mode"`
	AzimuthCommandedPosition                      float64 `json:"Azimuth commanded position"`     // Unit: [deg]
	AzimuthCurrentPosition                        float64 `json:"Azimuth current position"`       // Unit: [deg]
	AzimuthCurrentVelocity                        float64 `json:"Azimuth current velocity"`       // Unit: [deg/s]
	AzimuthAveragePositionError                   float64 `json:"Azimuth average position error"` // Unit: [deg]
	AzimuthPeakPositionError                      float64 `json:"Azimuth peak position error"`    // Unit: [deg]
	AzimuthComputerDisabled                       bool    `json:"Azimuth computer disabled"`
	AzimuthAxisDisabled                           bool    `json:"Azimuth axis disabled"`
	AzimuthAxisInStop                             bool    `json:"Azimuth axis in stop"`
	AzimuthBrakesReleased                         bool    `json:"Azimuth brakes released"`
	AzimuthStopAtLCP                              bool    `json:"Azimuth stop at LCP"`
	AzimuthPowerOn                                bool    `json:"Azimuth power on"`
	AzimuthCCWLimit2NdEmergency                   bool    `json:"Azimuth CCW limit: 2nd emergency"`
	AzimuthCCWLimitEmergency                      bool    `json:"Azimuth CCW limit: emergency"`
	AzimuthCCWLimitOperating                      bool    `json:"Azimuth CCW limit: operating"`
	AzimuthCCWLimitPreLimit                       bool    `json:"Azimuth CCW limit: pre-limit"`
	AzimuthCCWLimitOperatingAcuSoftwareLimit      bool    `json:"Azimuth CCW limit: operating (ACU software limit)"`
	AzimuthCCWLimitPreLimitAcuSoftwareLimit       bool    `json:"Azimuth CCW limit: pre-limit (ACU software limit)"`
	AzimuthCWLimitPreLimitAcuSoftwareLimit        bool    `json:"Azimuth CW limit: pre-limit (ACU software limit)"`
	AzimuthCWLimitOperatingAcuSoftwareLimit       bool    `json:"Azimuth CW limit: operating (ACU software limit)"`
	AzimuthCWLimitPreLimit                        bool    `json:"Azimuth CW limit: pre-limit"`
	AzimuthCWLimitOperating                       bool    `json:"Azimuth CW limit: operating"`
	AzimuthCWLimitEmergency                       bool    `json:"Azimuth CW limit: emergency"`
	AzimuthCWLimit2NdEmergency                    bool    `json:"Azimuth CW limit: 2nd emergency"`
	AzimuthSummaryFault                           bool    `json:"Azimuth summary fault"`
	AzimuthServoFailure                           bool    `json:"Azimuth servo failure"`
	AzimuthMotionError                            bool    `json:"Azimuth motion error"`
	AzimuthBrake1Failure                          bool    `json:"Azimuth brake 1 failure"`
	AzimuthBrake2Failure                          bool    `json:"Azimuth brake 2 failure"`
	AzimuthBreakerFailure                         bool    `json:"Azimuth breaker failure"`
	AzimuthAmplifier1Failure                      bool    `json:"Azimuth amplifier 1 failure"`
	AzimuthAmplifier2Failure                      bool    `json:"Azimuth amplifier 2 failure"`
	AzimuthMotor1Overtemperature                  bool    `json:"Azimuth motor 1 overtemperature"`
	AzimuthMotor2Overtemperature                  bool    `json:"Azimuth motor 2 overtemperature"`
	AzimuthAUX1ModeSelected                       bool    `json:"Azimuth AUX 1 mode selected"`
	AzimuthAUX2ModeSelected                       bool    `json:"Azimuth AUX 2 mode selected"`
	AzimuthOverspeed                              bool    `json:"Azimuth overspeed"`
	AzimuthAmplifierPowerCylceInterlock           bool    `json:"Azimuth amplifier power cylce interlock"`
	AzimuthRegenerationResistor1Overtemperature   bool    `json:"Azimuth regeneration resistor 1 overtemperature"`
	AzimuthRegenerationResistor2Overtemperature   bool    `json:"Azimuth regeneration resistor 2 overtemperature"`
	AzimuthCANBusAmplifier1CommunicationFailure   bool    `json:"Azimuth CAN bus amplifier 1 communication failure"`
	AzimuthCANBusAmplifier2CommunicationFailure   bool    `json:"Azimuth CAN bus amplifier 2 communication failure"`
	AzimuthEncoderFailure                         bool    `json:"Azimuth encoder failure"`
	AzimuthOscillationWarning                     bool    `json:"Azimuth oscillation warning"`
	AzimuthOscillationAlarm                       bool    `json:"Azimuth oscillation alarm"`
	AzimuthTachoFailure                           bool    `json:"Azimuth tacho failure"`
	AzimuthImmobile                               bool    `json:"Azimuth immobile"`
	AzimuthOvercurrentMotor1                      bool    `json:"Azimuth overcurrent motor 1"`
	AzimuthOvercurrentMotor2                      bool    `json:"Azimuth overcurrent motor 2"`
	ElevationMode                                 uint8   `json:"Elevation mode"`
	ElevationCommandedPosition                    float64 `json:"Elevation commanded position"`     // Unit: [deg]
	ElevationCurrentPosition                      float64 `json:"Elevation current position"`       // Unit: [deg]
	ElevationCurrentVelocity                      float64 `json:"Elevation current velocity"`       // Unit: [deg/s]
	ElevationAveragePositionError                 float64 `json:"Elevation average position error"` // Unit: [deg]
	ElevationPeakPositionError                    float64 `json:"Elevation peak position error"`    // Unit: [deg]
	ElevationComputerDisabled                     bool    `json:"Elevation computer disabled"`
	ElevationAxisDisabled                         bool    `json:"Elevation axis disabled"`
	ElevationAxisInStop                           bool    `json:"Elevation axis in stop"`
	ElevationBrakesReleased                       bool    `json:"Elevation brakes released"`
	ElevationStopAtLCP                            bool    `json:"Elevation stop at LCP"`
	ElevationPowerOn                              bool    `json:"Elevation power on"`
	ElevationDownLimitEmergency                   bool    `json:"Elevation Down limit: emergency"`
	ElevationDownLimitOperating                   bool    `json:"Elevation Down limit: operating"`
	ElevationDownLimitPreLimit                    bool    `json:"Elevation Down limit: pre-limit"`
	ElevationDownLimitOperatingAcuSoftwareLimit   bool    `json:"Elevation Down limit: operating (ACU software limit)"`
	ElevationDownLimitPreLimitAcuSoftwareLimit    bool    `json:"Elevation Down limit: pre-limit (ACU software limit)"`
	ElevationUpLimitPreLimitAcuSoftwareLimit      bool    `json:"Elevation Up limit: pre-limit (ACU software limit)"`
	ElevationUpLimitOperatingAcuSoftwareLimit     bool    `json:"Elevation Up limit: operating (ACU software limit)"`
	ElevationUpLimitPreLimit                      bool    `json:"Elevation Up limit: pre-limit"`
	ElevationUpLimitOperating                     bool    `json:"Elevation Up limit: operating"`
	ElevationUpLimitEmergency                     bool    `json:"Elevation Up limit: emergency"`
	ElevationSummaryFault                         bool    `json:"Elevation summary fault"`
	ElevationServoFailure                         bool    `json:"Elevation servo failure"`
	ElevationMotionError                          bool    `json:"Elevation motion error"`
	ElevationBrake1Failure                        bool    `json:"Elevation brake 1 failure"`
	ElevationBreakerFailure                       bool    `json:"Elevation breaker failure"`
	ElevationAmplifier1Failure                    bool    `json:"Elevation amplifier 1 failure"`
	ElevationMotor1Overtemp                       bool    `json:"Elevation motor 1 overtemp"`
	ElevationOverspeed                            bool    `json:"Elevation overspeed"`
	ElevationAmplifierPowerCylceInterlock         bool    `json:"Elevation amplifier power cylce interlock"`
	ElevationRegenerationResistor1Overtemperature bool    `json:"Elevation regeneration resistor 1 overtemperature"`
	ElevationCANBusAmplifier1CommunicationFailure bool    `json:"Elevation CAN bus amplifier 1 communication failure"`
	ElevationEncoderFailure                       bool    `json:"Elevation encoder failure"`
	ElevationOscillationWarning                   bool    `json:"Elevation oscillation warning"`
	ElevationOscillationAlarm                     bool    `json:"Elevation oscillation alarm"`
	ElevationImmobile                             bool    `json:"Elevation immobile"`
	ElevationOvercurrentMotor1                    bool    `json:"Elevation overcurrent motor 1"`
	BoresightMode                                 uint8   `json:"Boresight mode"`
	BoresightCommandedPosition                    float64 `json:"Boresight commanded position"` // Unit: [deg]
	BoresightCurrentPosition                      float64 `json:"Boresight current position"`   // Unit: [deg]
	BoresightComputerDisabled                     bool    `json:"Boresight computer disabled"`
	BoresightAxisDisabled                         bool    `json:"Boresight axis disabled"`
	BoresightAxisInStop                           bool    `json:"Boresight axis in stop"`
	BoresightBrakesReleased                       bool    `json:"Boresight brakes released"`
	BoresightStopAtLCP                            bool    `json:"Boresight stop at LCP"`
	BoresightPowerOn                              bool    `json:"Boresight power on"`
	BoresightCCWLimitEmergency                    bool    `json:"Boresight CCW limit: emergency"`
	BoresightCCWLimitOperating                    bool    `json:"Boresight CCW limit: operating"`
	BoresightCCWLimitPreLimit                     bool    `json:"Boresight CCW limit: pre-limit"`
	BoresightCCWLimitOperatingAcuSoftwareLimit    bool    `json:"Boresight CCW limit: operating (ACU software limit)"`
	BoresightCCWLimitPreLimitAcuSoftwareLimit     bool    `json:"Boresight CCW limit: pre-limit (ACU software limit)"`
	BoresightCWLimitPreLimitAcuSoftwareLimit      bool    `json:"Boresight CW limit: pre-limit (ACU software limit)"`
	BoresightCWLimitOperatingAcuSoftwareLimit     bool    `json:"Boresight CW limit: operating (ACU software limit)"`
	BoresightCWLimitPreLimit                      bool    `json:"Boresight CW limit: pre-limit"`
	BoresightCWLimitOperating                     bool    `json:"Boresight CW limit: operating"`
	BoresightCWLimitEmergency                     bool    `json:"Boresight CW limit: emergency"`
	BoresightSummaryFault                         bool    `json:"Boresight summary fault"`
	BoresightServoFailure                         bool    `json:"Boresight servo failure"`
	BoresightMotionError                          bool    `json:"Boresight motion error"`
	BoresightBrake1Failure                        bool    `json:"Boresight brake 1 failure"`
	BoresightBrake2Failure                        bool    `json:"Boresight brake 2 failure"`
	BoresightBreakerFailure                       bool    `json:"Boresight breaker failure"`
	BoresightAmplifier1Failure                    bool    `json:"Boresight amplifier 1 failure"`
	BoresightAmplifier2Failure                    bool    `json:"Boresight amplifier 2 failure"`
	BoresightMotor1Overtemperature                bool    `json:"Boresight motor 1 overtemperature"`
	BoresightMotor2Overtemperature                bool    `json:"Boresight motor 2 overtemperature"`
	BoresightAUX1ModeSelected                     bool    `json:"Boresight AUX 1 mode selected"`
	BoresightAUX2ModeSelected                     bool    `json:"Boresight AUX 2 mode selected"`
	BoresightOverspeed                            bool    `json:"Boresight overspeed"`
	BoresightAmplifierPowerCylceInterlock         bool    `json:"Boresight amplifier power cylce interlock"`
	BoresightRegenerationResistor1Overtemperature bool    `json:"Boresight regeneration resistor 1 overtemperature"`
	BoresightRegenerationResistor2Overtemperature bool    `json:"Boresight regeneration resistor 2 overtemperature"`
	BoresightCANBusAmplifier1CommunicationFailure bool    `json:"Boresight CAN bus amplifier 1 communication failure"`
	BoresightCANBusAmplifier2CommunicationFailure bool    `json:"Boresight CAN bus amplifier 2 communication failure"`
	BoresightEncoderFailure                       bool    `json:"Boresight encoder failure"`
	BoresightOscillationWarning                   bool    `json:"Boresight oscillation warning"`
	BoresightOscillationAlarm                     bool    `json:"Boresight oscillation alarm"`
	BoresightTachoFailure                         bool    `json:"Boresight tacho failure"`
	BoresightImmobile                             bool    `json:"Boresight immobile"`
	BoresightOvercurrentMotor1                    bool    `json:"Boresight overcurrent motor 1"`
	BoresightOvercurrentMotor2                    bool    `json:"Boresight overcurrent motor 2"`
	GeneralSummaryFault                           bool    `json:"General summary fault"`
	EStopServoDriveCabinet                        bool    `json:"E-Stop servo drive cabinet"`
	EStopServicePole                              bool    `json:"E-Stop service pole"`
	EStopAzMovable                                bool    `json:"E-Stop Az movable"`
	KeySwitchBypassEmergencyLimit                 bool    `json:"Key Switch Bypass Emergency Limit"`
	PCUOperation                                  bool    `json:"PCU operation"`
	Safe                                          bool    `json:"Safe"`
	PowerFailureLatched                           bool    `json:"Power failure (latched)"`
	LightningProtectionSurgeArresters             bool    `json:"Lightning protection surge arresters"`
	PowerFailureNotLatched                        bool    `json:"Power failure (not latched)"`
	PowerFailure24V                               bool    `json:"24V power failure"`
	GeneralBreakerFailure                         bool    `json:"General Breaker failure"`
	CabinetOvertemperature                        bool    `json:"Cabinet Overtemperature"`
	AmbientTemperatureLowOperationInhibited       bool    `json:"Ambient temperature low (operation inhibited)"`
	CoMovingShieldOff                             bool    `json:"Co-Moving Shield off"`
	PLCACUInterfaceError                          bool    `json:"PLC-ACU interface error"`
	QtyOfFreeProgramTrackStackPositions           int32   `json:"Qty of free program track stack positions"`
	ACUInRemoteMode                               bool    `json:"ACU in remote mode"`
	ACUFanFailure                                 bool    `json:"ACU fan failure"`
	CabinetUndertemperature                       bool    `json:"Cabinet undertemperature"`
	TimeSynchronisationError                      bool    `json:"Time synchronisation error"`
	ACUPLCCommunicationError                      bool    `json:"ACU-PLC communication error"`
	StartOfProgramTrackTooEarly                   bool    `json:"Start of Program Track too early"`
	TurnaroundAccelerationTooHigh                 bool    `json:"Turnaround acceleration too high"`
	TurnaroundTimeTooShort                        bool    `json:"Turnaround time too short"`
}
