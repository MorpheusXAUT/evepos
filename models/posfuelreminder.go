package models

import (
	"time"
)

type POSFuelReminder struct {
	Starbase     *POS
	ReminderTime time.Time
}

func NewPOSFuelReminder(starbase *POS) *POSFuelReminder {
	reminder := &POSFuelReminder{
		Starbase:     starbase,
		ReminderTime: time.Now(),
	}

	return reminder
}
