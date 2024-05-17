package datasets

type Status3rdAxis struct {
	ThirdAxisAxisInStowPosition bool  `json:"3rd axis axis in stow position"`
	ThirdAxisStowPinsStatus     uint8 `json:"3rd axis stow pins - status"`
	ThirdAxisSummaryFault       uint8 `json:"3rd axis summary fault"`
}
