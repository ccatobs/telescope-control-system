package datasets

type Shutter struct {
	ShutterClosed  bool `json:"Shutter Closed"`
	ShutterMoving  bool `json:"Shutter Moving"`
	ShutterOpen    bool `json:"Shutter Open"`
	ShutterTimeout bool `json:"Shutter Timeout"`
	ShutterFailure bool `json:"Shutter Failure"`
	MoveInterlock  bool `json:"Move Interlock"`
}
