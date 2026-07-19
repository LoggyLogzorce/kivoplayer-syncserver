package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Track struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Fingerprint string    `gorm:"uniqueIndex;not null"`
	Title       string
	Artist      string
	Album       string
	Duration    int // секунды
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Lyrics []Lyrics `gorm:"foreignKey:TrackID"`
}

type Lyrics struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TrackID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Source    string    // "embedded", "lrc", "manual"
	Content   string    `gorm:"type:text"`
	IsTimed   bool      // true = lrc формат с таймкодами
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Track) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (l *Lyrics) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
