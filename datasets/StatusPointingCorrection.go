package datasets

type StatusPointingCorrection struct {
	Time                       float64 `json:"Time"`
	Year                       uint32  `json:"Year"`
	SystematicErrorModelSpemOn bool    `json:"Systematic error model (SPEM) on"`
	TiltmeterCorrectionOn      bool    `json:"Tiltmeter correction on"`
	LinearSensorCorrectionOn   bool    `json:"Linear sensor correction on"`
	RFRefractionCorrectionOn   bool    `json:"RF refraction correction on"`
	Spare4                     bool    `json:"Spare 4"`
	Spare5                     bool    `json:"Spare 5"`
	Spare6                     bool    `json:"Spare 6"`
	Spare7                     bool    `json:"Spare 7"`
	Spare8                     bool    `json:"Spare 8"`
	Spare9                     bool    `json:"Spare 9"`
	SPEMCorrectionAZ           float64 `json:"SPEM correction AZ"` // Unit: [deg]
	SPEMCorrectionEL           float64 `json:"SPEM correction EL"` // Unit: [deg]
	TiltmeterCorrectionAZ      float64 `json:"Tiltmeter correction AZ"`
	TiltmeterCorrectionEL      float64 `json:"Tiltmeter correction EL"`
	LinearSensorCorrectionAZ   float64 `json:"Linear sensor correction AZ"`
	LinearSensorCorrectionEL   float64 `json:"Linear sensor correction EL"`
	RFRefractionCorrectionAZ   float64 `json:"RF refraction correction AZ"`
	RFRefractionCorrectionEL   float64 `json:"RF refraction correction EL"` // Unit: [deg]
}
