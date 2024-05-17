package datasets

type CmdStarTrackTransfer struct {
	Name           [32]byte `json:"Name"`
	RightAscension float64  `json:"Right Ascension"`
	Declination    float64  `json:"Declination"`
}
