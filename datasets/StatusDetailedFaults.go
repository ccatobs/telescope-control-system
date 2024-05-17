package datasets

type StatusDetailedFaults struct {
	PLCCommunicationTcp                       uint8 `json:"PLC Communication (TCP)"`
	EStop1ServoDriveCabinet                   uint8 `json:"EStop 1 - Servo Drive Cabinet"`
	EStop2AZDrives12                          uint8 `json:"EStop 2 - AZ Drives 1+2"`
	EStop3AzDrives34                          uint8 `json:"EStop 3 - Az Drives 3+4"`
	EStop4ElDrives                            uint8 `json:"EStop 4 - El Drives"`
	EStop5StaircaseLowerEnd                   uint8 `json:"EStop 5 - Staircase Lower End"`
	EStop6ElevatorAccess                      uint8 `json:"EStop 6 - Elevator Access"`
	EStop7InstrumentSpace1CoRotator           uint8 `json:"EStop 7 - Instrument Space 1, Co-Rotator"`
	EStop8ELHousingMirrorArea                 uint8 `json:"EStop 8 - EL Housing, Mirror Area"`
	EStop9MPD                                 uint8 `json:"EStop 9 - MPD"`
	EStop10OCS                                uint8 `json:"EStop 10 - OCS"`
	EStopPCU                                  uint8 `json:"E-Stop PCU"`
	EStopDevice                               uint8 `json:"E-Stop Device"`
	DriveCabinetDoor                          uint8 `json:"Drive cabinet door"`
	HoistStorageContainer                     uint8 `json:"Hoist storage container"`
	AccessHatchInterlockSupportCone           uint8 `json:"Access Hatch Interlock - Support Cone"`
	YokeADoorWarningOutsideToStairway         uint8 `json:"Yoke A Door Warning - Outside to Stairway"`
	StairwayDoorWarningYokeTraverse           uint8 `json:"Stairway Door Warning - Yoke Traverse"`
	YokeTraverseElectronicsSpaceHoistPosition uint8 `json:"Yoke Traverse: Electronics space hoist position"`
	LadderInterlockYokeTraverse               uint8 `json:"Ladder Interlock - Yoke Traverse"`
	YokeADoorInterlock3RdFloor                uint8 `json:"Yoke A Door Interlock - 3rd Floor"`
	FloorHatchInterlockInstrumentSpace1       uint8 `json:"Floor Hatch Interlock - Instrument Space 1"`
	MaterialDoorWarningInstrumentSpace        uint8 `json:"Material Door Warning - Instrument Space"`
	YokeADoorInterlock2NdFloor                uint8 `json:"Yoke A Door Interlock - 2nd Floor"`
	YokeBDoorWarningTraverseTo1StFloor        uint8 `json:"Yoke B Door Warning - Traverse to 1st Floor"`
	AccessHatchInterlockRoofYokeB             uint8 `json:"Access Hatch Interlock - Roof Yoke B"`
	CraneOn                                   uint8 `json:"Crane on"`
	PLCGeneralSafe                            uint8 `json:"PLC General - Safe"`
	PLCGeneralPowerFailureLatched             uint8 `json:"PLC General - Power Failure (latched)"`
	PLCGeneralLightningSurgeArrester          uint8 `json:"PLC General - Lightning Surge Arrester"`
	PLCGeneralPowerFailure                    uint8 `json:"PLC General - Power Failure"`
	PLCGeneral24VPowerFailure                 uint8 `json:"PLC General - 24V Power Failure"`
	PLCGeneralBreakerFailure                  uint8 `json:"PLC General - Breaker Failure"`
	PLCGeneralCabinetOverTemperature          uint8 `json:"PLC General - Cabinet Over-Temperature"`
	PLCGeneralACUIFError                      uint8 `json:"PLC General - ACU IF Error"`
	PLCGeneralProfinetError                   uint8 `json:"PLC General - Profinet Error"`
	PLCGeneralCabinetUnderTemperature         uint8 `json:"PLC General - Cabinet Under-Temperature"`
	PLCGeneralAmbientTemperatureTooLow        uint8 `json:"PLC General - Ambient-Temperature too low"`
	PLCGeneralComovingShieldRemoved           uint8 `json:"PLC General - Comoving Shield removed"`
	PLCGeneralFanFailure                      uint8 `json:"PLC General - Fan Failure"`
	PLCGeneralDrivePowerOff                   uint8 `json:"PLC General - Drive Power Off"`
	PLCGeneralKeySwitchSafeOverride           uint8 `json:"PLC General - Key Switch Safe Override"`
	PLCGeneralKeySwitchBypassEmergencyLimits  uint8 `json:"PLC General - Key Switch Bypass Emergency Limits"`
	PLCGeneralPCUOperation                    uint8 `json:"PLC General - PCU Operation"`
	PLCGeneralWarnHornSounding                uint8 `json:"PLC General - Warn Horn sounding"`
	PLCGeneralWarnHornPassive                 uint8 `json:"PLC General - Warn Horn passive"`
	AzEncoderFailure                          uint8 `json:"Az - Encoder Failure"`
	AzRoundSwitchFailure                      uint8 `json:"Az - Round Switch Failure"`
	AzSecondEmergencyLimitUpCW                uint8 `json:"Az - Second Emergency Limit Up/CW"`
	AzEmergencyLimitUpCW                      uint8 `json:"Az - Emergency Limit Up/CW"`
	AzLimitUpCW                               uint8 `json:"Az - Limit Up/CW"`
	AzPreLimitUpCW                            uint8 `json:"Az - PreLimit Up/CW"`
	AzPreLimitDownCCW                         uint8 `json:"Az - PreLimit Down/CCW"`
	AzLimitDownCCW                            uint8 `json:"Az - Limit Down/CCW"`
	AzEmergencyLimitDownCCW                   uint8 `json:"Az - Emergency Limit Down/CCW"`
	AzSecondEmergencyLimitDownCCW             uint8 `json:"Az - Second Emergency Limit Down/CCW"`
	AzSoftLimitUpCW                           uint8 `json:"Az - Soft-Limit Up/CW"`
	AzSoftPreLimitUpCW                        uint8 `json:"Az - Soft-PreLimit Up/CW"`
	AzSoftPreLimitDownCCW                     uint8 `json:"Az - Soft-PreLimit Down/CCW"`
	AzSoftLimitDownCCW                        uint8 `json:"Az - Soft-Limit Down/CCW"`
	AzTachoFailure                            uint8 `json:"Az - Tacho Failure"`               // Unit: [BitFault]
	AzACUImmobile                             uint8 `json:"Az - ACU immobile"`                // Unit: [BitFault]
	AzCANBusFailureAmplifier1                 uint8 `json:"Az - CAN bus failure Amplifier 1"` // Unit: [BitFault]
	AzCANBusFailureAmplifier2                 uint8 `json:"Az - CAN bus failure Amplifier 2"` // Unit: [BitFault]
	AzCANBusFailureAmplifier3                 uint8 `json:"Az - CAN bus failure Amplifier 3"` // Unit: [BitFault]
	AzCANBusFailureAmplifier4                 uint8 `json:"Az - CAN bus failure Amplifier 4"` // Unit: [BitFault]
	AzComputerDisabled                        uint8 `json:"Az - Computer Disabled"`
	AzAxisDisabled                            uint8 `json:"Az - Axis Disabled"`
	AzServoFailure                            uint8 `json:"Az - Servo Failure"`
	AzMotionError                             uint8 `json:"Az - Motion Error"`
	AzBrake1Failure                           uint8 `json:"Az - Brake1 Failure"`
	AzBrake2Failure                           uint8 `json:"Az - Brake2 Failure"`
	AzBrake3Failure                           uint8 `json:"Az - Brake3 Failure"`
	AzBrake4Failure                           uint8 `json:"Az - Brake4 Failure"`
	AzBreakerFailure                          uint8 `json:"Az - Breaker Failure"`
	AzAmplifier1Failure                       uint8 `json:"Az - Amplifier1 Failure"`
	AzAmplifier2Failure                       uint8 `json:"Az - Amplifier2 Failure"`
	AzAmplifier3Failure                       uint8 `json:"Az - Amplifier3 Failure"`
	AzAmplifier4Failure                       uint8 `json:"Az - Amplifier4 Failure"`
	AzMotor1OverTemp                          uint8 `json:"Az - Motor1 OverTemp"`
	AzMotor2OverTemp                          uint8 `json:"Az - Motor2 OverTemp"`
	AzMotor3OverTemp                          uint8 `json:"Az - Motor3 OverTemp"`
	AzMotor4OverTemp                          uint8 `json:"Az - Motor4 OverTemp"`
	AzAux1ModeSelected                        uint8 `json:"Az - Aux1 Mode Selected"`
	AzAux2ModeSelected                        uint8 `json:"Az - Aux2 Mode Selected"`
	AzOverspeed                               uint8 `json:"Az - Overspeed"`
	AzGearbox1LowOilLevel                     uint8 `json:"Az - Gearbox1 Low Oil Level"`
	AzGearbox1HeaterFailure                   uint8 `json:"Az - Gearbox1 Heater Failure"`
	AzGearbox2LowOilLevel                     uint8 `json:"Az - Gearbox2 Low Oil Level"`
	AzGearbox2HeaterFailure                   uint8 `json:"Az - Gearbox2 Heater Failure"`
	AzGearbox3LowOilLevel                     uint8 `json:"Az - Gearbox3 Low Oil Level"`
	AzGearbox3HeaterFailure                   uint8 `json:"Az - Gearbox3 Heater Failure"`
	AzGearbox4LowOilLevel                     uint8 `json:"Az - Gearbox4 Low Oil Level"`
	AzGearbox4HeaterFailure                   uint8 `json:"Az - Gearbox4 Heater Failure"`
	AzAmplifierPowerCycleInterlock            uint8 `json:"Az - Amplifier Power Cycle Interlock"`
	AzDCBus1Failure                           uint8 `json:"Az - DC Bus1 Failure"`
	AzDCBus2Failure                           uint8 `json:"Az - DC Bus2 Failure"`
	AzSecondaryEncoderFailure                 uint8 `json:"Az - Secondary Encoder Failure"`
	AzRegenResistor1OverTemp                  uint8 `json:"Az - Regen. Resistor1 OverTemp"`
	AzRegenResistor2OverTemp                  uint8 `json:"Az - Regen. Resistor2 OverTemp"`
	AzRegenResistor3OverTemp                  uint8 `json:"Az - Regen. Resistor3 OverTemp"`
	AzRegenResistor4OverTemp                  uint8 `json:"Az - Regen. Resistor4 OverTemp"`
	ElEncoderFailure                          uint8 `json:"El - Encoder Failure"`
	ElSecondEmergencyLimitUpCW                uint8 `json:"El - Second Emergency Limit Up/CW"`
	ElEmergencyLimitUpCW                      uint8 `json:"El - Emergency Limit Up/CW"`
	ElLimitUpCW                               uint8 `json:"El - Limit Up/CW"`
	ElPreLimitUpCW                            uint8 `json:"El - PreLimit Up/CW"`
	ElPreLimitDownCCW                         uint8 `json:"El - PreLimit Down/CCW"`
	ElLimitDownCCW                            uint8 `json:"El - Limit Down/CCW"`
	ElEmergencyLimitDownCCW                   uint8 `json:"El - Emergency Limit Down/CCW"`
	ElSecondEmergencyLimitDownCCW             uint8 `json:"El - Second Emergency Limit Down/CCW"`
	ElSoftLimitUpCW                           uint8 `json:"El - Soft-Limit Up/CW"`
	ElSoftPreLimitUpCW                        uint8 `json:"El - Soft-PreLimit Up/CW"`
	ElSoftPreLimitDownCCW                     uint8 `json:"El - Soft-PreLimit Down/CCW"`
	ElSoftLimitDownCCW                        uint8 `json:"El - Soft-Limit Down/CCW"`
	ElTachoFailure                            uint8 `json:"El - Tacho Failure"`               // Unit: [BitFault]
	ElACUImmobile                             uint8 `json:"El - ACU immobile"`                // Unit: [BitFault]
	ElCANBusFailureAmplifier1                 uint8 `json:"El - CAN bus failure Amplifier 1"` // Unit: [BitFault]
	ElCANBusFailureAmplifier2                 uint8 `json:"El - CAN bus failure Amplifier 2"` // Unit: [BitFault]
	ElComputerDisabled                        uint8 `json:"El - Computer Disabled"`
	ElAxisDisabled                            uint8 `json:"El - Axis Disabled"`
	ElServoFailure                            uint8 `json:"El - Servo Failure"`
	ElMotionError                             uint8 `json:"El - Motion Error"`
	ElBrake1Failure                           uint8 `json:"El - Brake1 Failure"`
	ElBrake2Failure                           uint8 `json:"El - Brake2 Failure"`
	ElBreakerFailure                          uint8 `json:"El - Breaker Failure"`
	ElAmplifier1Failure                       uint8 `json:"El - Amplifier1 Failure"`
	ElAmplifier2Failure                       uint8 `json:"El - Amplifier2 Failure"`
	ElMotor1OverTemp                          uint8 `json:"El - Motor1 OverTemp"`
	ElMotor2OverTemp                          uint8 `json:"El - Motor2 OverTemp"`
	ElAux1ModeSelected                        uint8 `json:"El - Aux1 Mode Selected"`
	ElAux2ModeSelected                        uint8 `json:"El - Aux2 Mode Selected"`
	ElOverspeed                               uint8 `json:"El - Overspeed"`
	ElGearbox1LowOilLevel                     uint8 `json:"El - Gearbox1 Low Oil Level"`
	ElGearbox1HeaterFailure                   uint8 `json:"El - Gearbox1 Heater Failure"`
	ElGearbox2LowOilLevel                     uint8 `json:"El - Gearbox2 Low Oil Level"`
	ElGearbox2HeaterFailure                   uint8 `json:"El - Gearbox2 Heater Failure"`
	ElAmplifierPowerCycleInterlock            uint8 `json:"El - Amplifier Power Cycle Interlock"`
	ElSecondaryEncoderFailure                 uint8 `json:"El - Secondary Encoder Failure"`
	ElRegenResistor1OverTemp                  uint8 `json:"El - Regen. Resistor1 OverTemp"`
	ElRegenResistor2OverTemp                  uint8 `json:"El - Regen. Resistor2 OverTemp"`
	GeneralACUFanFailure                      uint8 `json:"General ACU fan failure"`
	GeneralACUOverTemperature                 uint8 `json:"General ACU over temperature"`
	GeneralTimeError                          uint8 `json:"General Time Error"` // Unit: [BitFault]
}
