package model

import (
	"gorm.io/gorm"
	"nuchal-api/db"
)

type Posture struct {
	gorm.Model

	// PatternID is the primary
	PatternID string `json:"pattern_id"`

	// Enable is a flag used to determine whether it is used in live trading.
	Enable bool `json:"enable"`

	// Enter is the unix time stamp for entering a 24-hour trading period.
	Enter int64 `json:"enter"`

	// Exit is the unix time stamp for exiting a 24-hour trading period.
	Exit int64 `json:"exit"`
}

func init() {
	db.Migrate(&Posture{})
}
