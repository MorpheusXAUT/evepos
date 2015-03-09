package models

import (
	"encoding/json"
)

// POS represents a player operated starbase
type POS struct {
	// ID represents the database ID of the POS
	ID int64 `json:"id"`
	// Active indicates whether the User is set as active
	Active bool `json:"active"`
}

// NewPOS creates a new POS with the given information
func NewPOS(active bool) *POS {
	pos := &POS{
		ID:     -1,
		Active: active,
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
