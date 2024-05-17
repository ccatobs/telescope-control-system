package datasets

type CmdPointingCorrection struct {
	Time                       float64 `json:"Time"`
	Year                       uint32  `json:"Year"`
	SystematicErrorModelSpemOn bool    `json:"Systematic error model (SPEM) on"`
	TiltmeterAzCorrectionOn    bool    `json:"Tiltmeter Az correction on"`
	TiltmeterElCorrectionOn    bool    `json:"Tiltmeter El correction on"`
	RFRefractionCorrectionOn   bool    `json:"RF refraction correction on"`
	Spare5                     bool    `json:"Spare 5"`
	Spare6                     bool    `json:"Spare 6"`
	Spare7                     bool    `json:"Spare 7"`
	Spare8                     bool    `json:"Spare 8"`
	SPEMCorrectionAZ           float64 `json:"SPEM correction AZ"`          // Unit: [deg]
	SPEMCorrectionEL           float64 `json:"SPEM correction EL"`          // Unit: [deg]
	TiltmeterAzCorrectionAZ    float64 `json:"Tiltmeter  Az correction AZ"` // Unit: [deg]
	TiltmeterAzCorrectionEL    float64 `json:"Tiltmeter  Az correction EL"` // Unit: [deg]
	TiltmeterElCorrectionAZ    float64 `json:"Tiltmeter El correction AZ"`  // Unit: [deg]
	TiltmeterElCorrectionEL    float64 `json:"Tiltmeter El correction EL"`  // Unit: [deg]
	RFRefractionCorrectionAZ   float64 `json:"RF refraction correction AZ"`
	RFRefractionCorrectionEL   float64 `json:"RF refraction correction EL"` // Unit: [deg]
}
