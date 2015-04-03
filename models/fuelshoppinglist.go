package models

type FuelShoppingList struct {
	FuelList []*Fuel
}

type Fuel struct {
	TypeID   int64
	Name     string
	Quantity int64
	Volume   int64
}

func NewFuelShoppingList(fuelList []*Fuel) *FuelShoppingList {
	fuelShoppingList := &FuelShoppingList{
		FuelList: fuelList,
	}

	return fuelShoppingList
}

func NewFuel(typeID int64, name string, quantity int64) *Fuel {
	fuel := &Fuel{
		TypeID:   typeID,
		Name:     name,
		Quantity: quantity,
		Volume:   quantity * 5,
	}

	return fuel
}

func (fuelShoppingList *FuelShoppingList) CalculateTotalVolume() int64 {
	var totalVolume int64
	totalVolume = 0

	for _, fuel := range fuelShoppingList.FuelList {
		totalVolume += fuel.Volume
	}

	return totalVolume
}
