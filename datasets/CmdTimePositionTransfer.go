package datasets

type CmdTimePositionTransfer struct {
	LastUploadName           [48]byte `json:"Last Upload Name"`
	LastUploadTime           [32]byte `json:"Last Upload Time"`
	LastUploadTotalEntries   int32    `json:"Last Upload Total Entries"`
	LastUploadValidEntries   int32    `json:"Last Upload Valid Entries"`
	LastUploadInvalidEntries int32    `json:"Last Upload Invalid Entries"`
}
