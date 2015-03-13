package models

import (
	"encoding/json"

	"github.com/morpheusxaut/eveapi"
)

// POS represents a player operated starbase
type POS struct {
	Base    *eveapi.Starbase
	Details *eveapi.StarbaseDetails
	Fuel    *POSFuel
}

// NewPOS creates a new POS with the given information
func NewPOS(base *eveapi.Starbase, details *eveapi.StarbaseDetails, fuel *POSFuel) *POS {
	pos := &POS{
		Base:    base,
		Details: details,
		Fuel:    fuel,
	}

	return pos
}

// String represents a JSON encoded representation of the POS
func (pos *POS) String() string {
	jsonContent, err := json.Marshal(pos)
	if err != nil {
		return ""
	}

	return string(jsonContent)
}
