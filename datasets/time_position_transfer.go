package datasets

type TimePositionTransfer struct {
	Day        int32
	TimeOfDay  float64
	AzPosition float64
	ElPosition float64
	AzVelocity float64
	ElVelocity float64
	AzFlag     int8
	ElFlag     int8
}
