package datasets

type StatusCCATDetailed8100 struct {
	Time                                          float64 `json:"Time"`
	Year                                          uint32  `json:"Year"`
	AzimuthMode                                   uint8   `json:"Azimuth mode"`
	AzimuthCommandedPosition                      float64 `json:"Azimuth commanded position"`     // Unit: [deg]
	AzimuthCurrentPosition                        float64 `json:"Azimuth current position"`       // Unit: [deg]
	AzimuthCurrentVelocity                        float64 `json:"Azimuth current velocity"`       // Unit: [deg/s]
	AzimuthAveragePositionError                   float64 `json:"Azimuth average position error"` // Unit: [deg]
	AzimuthPeakPositionError                      float64 `json:"Azimuth peak position error"`    // Unit: [deg]
	AzimuthComputerDisabled                       bool    `json:"Azimuth computer disabled"`
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
	AzimuthBrake3Failure                          bool    `json:"Azimuth brake 3 failure"`
	AzimuthBrake4Failure                          bool    `json:"Azimuth brake 4 failure"`
	AzimuthBreakerFailure                         bool    `json:"Azimuth breaker failure"`
	AzimuthAmplifier1Failure                      bool    `json:"Azimuth amplifier 1 failure"`
	AzimuthAmplifier2Failure                      bool    `json:"Azimuth amplifier 2 failure"`
	AzimuthAmplifier3Failure                      bool    `json:"Azimuth amplifier 3 failure"`
	AzimuthAmplifier4Failure                      bool    `json:"Azimuth amplifier 4 failure"`
	AzimuthMotor1Overtemperature                  bool    `json:"Azimuth motor 1 overtemperature"`
	AzimuthMotor2Overtemperature                  bool    `json:"Azimuth motor 2 overtemperature"`
	AzimuthMotor3Overtemperature                  bool    `json:"Azimuth motor 3 overtemperature"`
	AzimuthMotor4Overtemperature                  bool    `json:"Azimuth motor 4 overtemperature"`
	AzimuthAUX1ModeSelected                       bool    `json:"Azimuth AUX 1 mode selected"`
	AzimuthAUX2ModeSelected                       bool    `json:"Azimuth AUX 2 mode selected"`
	AzimuthOverspeed                              bool    `json:"Azimuth overspeed"`
	AzimuthGearbox1LowOilLevel                    bool    `json:"Azimuth gearbox 1 low oil level"`
	AzimuthGearbox2LowOilLevel                    bool    `json:"Azimuth gearbox 2 low oil level"`
	AzimuthGearbox3LowOilLevel                    bool    `json:"Azimuth gearbox 3 low oil level"`
	AzimuthGearbox4LowOilLevel                    bool    `json:"Azimuth gearbox 4 low oil level"`
	AzimuthDCBus1Failure                          bool    `json:"Azimuth DC bus 1 failure"`
	AzimuthDCBus2Failure                          bool    `json:"Azimuth DC bus 2 failure"`
	AzimuthRegenerationResistor1Overtemperature   bool    `json:"Azimuth regeneration resistor 1 overtemperature"`
	AzimuthRegenerationResistor2Overtemperature   bool    `json:"Azimuth regeneration resistor 2 overtemperature"`
	AzimuthRegenerationResistor3Overtemperature   bool    `json:"Azimuth regeneration resistor 3 overtemperature"`
	AzimuthRegenerationResistor4Overtemperature   bool    `json:"Azimuth regeneration resistor 4 overtemperature"`
	AzimuthSecondaryEncoderFailure                bool    `json:"Azimuth Secondary Encoder Failure"`
	AzimuthCANBusAmplifier1CommunicationFailure   bool    `json:"Azimuth CAN bus amplifier 1 communication failure"`
	AzimuthCANBusAmplifier2CommunicationFailure   bool    `json:"Azimuth CAN bus amplifier 2 communication failure"`
	AzimuthCANBusAmplifier3CommunicationFailure   bool    `json:"Azimuth CAN bus amplifier 3 communication failure"`
	AzimuthCANBusAmplifier4CommunicationFailure   bool    `json:"Azimuth CAN bus amplifier 4 communication failure"`
	AzimuthEncoderFailure                         bool    `json:"Azimuth encoder failure"`
	AzimuthOscillationWarning                     bool    `json:"Azimuth oscillation warning"`
	AzimuthOscillationAlarm                       bool    `json:"Azimuth oscillation alarm"`
	AzimuthTachoFailure                           bool    `json:"Azimuth tacho failure"`
	AzimuthImmobile                               bool    `json:"Azimuth immobile"`
	AzimuthOvercurrentMotor1                      bool    `json:"Azimuth overcurrent motor 1"`
	AzimuthOvercurrentMotor2                      bool    `json:"Azimuth overcurrent motor 2"`
	AzimuthOvercurrentMotor3                      bool    `json:"Azimuth overcurrent motor 3"`
	AzimuthOvercurrentMotor4                      bool    `json:"Azimuth overcurrent motor 4"`
	ElevationMode                                 uint8   `json:"Elevation mode"`
	ElevationCommandedPosition                    float64 `json:"Elevation commanded position"`     // Unit: [deg]
	ElevationCurrentPosition                      float64 `json:"Elevation current position"`       // Unit: [deg]
	ElevationCurrentVelocity                      float64 `json:"Elevation current velocity"`       // Unit: [deg/s]
	ElevationAveragePositionError                 float64 `json:"Elevation average position error"` // Unit: [deg]
	ElevationPeakPositionError                    float64 `json:"Elevation peak position error"`    // Unit: [deg]
	ElevationComputerDisabled                     bool    `json:"Elevation computer disabled"`
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
	ElevationBrake2Failure                        bool    `json:"Elevation brake 2 failure"`
	ElevationBreakerFailure                       bool    `json:"Elevation breaker failure"`
	ElevationAmplifier1Failure                    bool    `json:"Elevation amplifier 1 failure"`
	ElevationAmplifier2Failure                    bool    `json:"Elevation amplifier 2 failure"`
	ElevationMotor1Overtemperature                bool    `json:"Elevation motor 1 overtemperature"`
	ElevationMotor2Overtemperature                bool    `json:"Elevation motor 2 overtemperature"`
	ElevationAUX1ModeSelected                     bool    `json:"Elevation AUX 1 mode selected"`
	ElevationAUX2ModeSelected                     bool    `json:"Elevation AUX 2 mode selected"`
	ElevationOverspeed                            bool    `json:"Elevation overspeed"`
	ElevationGearbox1LowOilLevel                  bool    `json:"Elevation gearbox 1 low oil level"`
	ElevationGearbox2LowOilLevel                  bool    `json:"Elevation gearbox 2 low oil level"`
	ElevationRegenerationResistor1Overtemperature bool    `json:"Elevation regeneration resistor 1 overtemperature"`
	ElevationRegenerationResistor2Overtemperature bool    `json:"Elevation regeneration resistor 2 overtemperature"`
	ElevationSecondaryEncoderFailure              bool    `json:"Elevation Secondary Encoder Failure"`
	ElevationCANBusAmplifier1CommunicationFailure bool    `json:"Elevation CAN bus amplifier 1 communication failure"`
	ElevationCANBusAmplifier2CommunicationFailure bool    `json:"Elevation CAN bus amplifier 2 communication failure"`
	ElevationEncoderFailure                       bool    `json:"Elevation encoder failure"`
	ElevationOscillationWarning                   bool    `json:"Elevation oscillation warning"`
	ElevationOscillationAlarm                     bool    `json:"Elevation oscillation alarm"`
	ElevationImmobile                             bool    `json:"Elevation immobile"`
	ElevationOvercurrentMotor1                    bool    `json:"Elevation overcurrent motor 1"`
	ElevationOvercurrentMotor2                    bool    `json:"Elevation overcurrent motor 2"`
	GeneralSummaryFault                           bool    `json:"General summary fault"`
	EStopServoDriveCabinet                        bool    `json:"E-Stop servo drive cabinet"`
	EStopAZDrives12                               bool    `json:"E-Stop AZ Drives 1+2"`
	EStopAzDrives34                               bool    `json:"E-Stop Az Drives 3+4"`
	EStopElDrives                                 bool    `json:"E-Stop El Drives"`
	EStopStaircaseLowerEnd                        bool    `json:"E-Stop Staircase Lower End"`
	EStopElevatorAccess                           bool    `json:"E-Stop Elevator Access"`
	EStopInstrumentSpace1CoRotator                bool    `json:"E-Stop Instrument Space 1, Co-Rotator"`
	EStopELHousingMirrorArea                      bool    `json:"E-Stop EL Housing, Mirror Area"`
	EStopMPD                                      bool    `json:"E-Stop MPD"`
	EStopOCS                                      bool    `json:"E-Stop OCS"`
	AccessHatchInterlockSupportCone               bool    `json:"Access Hatch Interlock - Support Cone"`
	YokeADoorWaringOutsideToStairway              bool    `json:"Yoke A Door Waring - Outside to Stairway"`
	KeySwitchSafeOverride                         bool    `json:"Key Switch Safe Override"`
	KeySwitchBypassEmergencyLimit                 bool    `json:"Key Switch Bypass Emergency Limit"`
	PCUOperation                                  bool    `json:"PCU operation"`
	Safe                                          bool    `json:"Safe"`
	PowerFailureLatched                           bool    `json:"Power failure (latched)"`
	LightningProtectionSurgeArresters             bool    `json:"Lightning protection surge arresters"`
	PowerFailureNotLatched                        bool    `json:"Power failure (not latched)"`
	PowerFailure24V                               bool    `json:"24V power failure"`
	GeneralBreakerFailure                         bool    `json:"General Breaker failure"`
	CabinetOvertemperature                        bool    `json:"Cabinet Overtemperature"`
	CabinetUndertemperature                       bool    `json:"Cabinet undertemperature"`
	ProfinetError                                 bool    `json:"Profinet Error"`
	AmbientTemperatureLowOperationInhibited       bool    `json:"Ambient temperature low (operation inhibited)"`
	CraneOn                                       bool    `json:"Crane on"`
	PLCACUInterfaceError                          bool    `json:"PLC-ACU interface error"`
	QtyOfFreeProgramTrackStackPositions           int32   `json:"Qty of free program track stack positions"`
	ACUInRemoteMode                               bool    `json:"ACU in remote mode"`
	ACUFanFailure                                 bool    `json:"ACU fan failure"`
	TimeSynchronisationError                      bool    `json:"Time synchronisation error"`
	ACUPLCCommunicationError                      bool    `json:"ACU-PLC communication error"`
}
