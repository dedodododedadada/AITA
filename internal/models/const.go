package models

import "time"

const (
	MaxSessionLife		= 7 * 24 * time.Hour
	SessionDuration     = 24 * time.Hour
	editWindow          = 10 * time.Minute
)