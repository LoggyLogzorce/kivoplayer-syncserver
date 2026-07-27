package models

import (
	"time"

	"github.com/google/uuid"
)

type Track struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Fingerprint string    `gorm:"uniqueIndex;not null"`
	FileName    string
	Duration    int // секунды
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Lyrics Lyrics `gorm:"foreignKey:TrackID"`
}

type Lyrics struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TrackID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Fingerprint string    `gorm:"uniqueIndex; not null"`
	FileName    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
