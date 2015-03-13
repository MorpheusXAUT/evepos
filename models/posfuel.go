package models

type POSFuel struct {
	TypeID   int64
	TypeName string
	Usage    int64
	Quantity int64
}

func NewPOSFuel(typeID int64, typeName string, usage int64, quantity int64) *POSFuel {
	fuel := &POSFuel{
		TypeID:   typeID,
		TypeName: typeName,
		Usage:    usage,
		Quantity: quantity,
	}

	return fuel
}
