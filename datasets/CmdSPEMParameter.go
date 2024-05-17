package datasets

type CmdSPEMParameter struct {
	Time          float64 `json:"Time"`
	Year          uint32  `json:"Year"`
	ParameterIA   float64 `json:"Parameter IA"`   // Unit: [mdeg]
	ParameterIE   float64 `json:"Parameter IE"`   // Unit: [mdeg]
	ParameterTFC  float64 `json:"Parameter TFC"`  // Unit: [mdeg]
	ParameterTF   float64 `json:"Parameter TF"`   // Unit: [mdeg]
	ParameterTFS  float64 `json:"Parameter TFS"`  // Unit: [mdeg]
	ParameterTFEC float64 `json:"Parameter TFEC"` // Unit: [mdeg]
	ParameterTFE  float64 `json:"Parameter TFE"`  // Unit: [mdeg]
	ParameterTFES float64 `json:"Parameter TFES"` // Unit: [mdeg]
	ParameterAN   float64 `json:"Parameter AN"`   // Unit: [mdeg]
	ParameterAW   float64 `json:"Parameter AW"`   // Unit: [mdeg]
	ParameterAN2  float64 `json:"Parameter AN2"`  // Unit: [mdeg]
	ParameterAW2  float64 `json:"Parameter AW2"`  // Unit: [mdeg]
	ParameterCA   float64 `json:"Parameter CA"`   // Unit: [mdeg]
	ParameterNPAE float64 `json:"Parameter NPAE"` // Unit: [mdeg]
	ParameterAES  float64 `json:"Parameter AES"`  // Unit: [mdeg]
	ParameterAEC  float64 `json:"Parameter AEC"`  // Unit: [mdeg]
	ParameterAES2 float64 `json:"Parameter AES2"` // Unit: [mdeg]
	ParameterAEC2 float64 `json:"Parameter AEC2"` // Unit: [mdeg]
	ParameterEES  float64 `json:"Parameter EES"`  // Unit: [mdeg]
	ParameterEEC  float64 `json:"Parameter EEC"`  // Unit: [mdeg]
	ParameterEES2 float64 `json:"Parameter EES2"` // Unit: [mdeg]
	ParameterEEC2 float64 `json:"Parameter EEC2"` // Unit: [mdeg]
	ParameterNRX  float64 `json:"Parameter NRX"`  // Unit: [mdeg]
	ParameterNRY  float64 `json:"Parameter NRY"`  // Unit: [mdeg]
	ParameterLELA float64 `json:"Parameter LELA"` // Unit: [mdeg]
	ParameterLELE float64 `json:"Parameter LELE"` // Unit: [mdeg]
	LELVelocity   float64 `json:"LEL velocity"`   // Unit: [deg/s]
}
