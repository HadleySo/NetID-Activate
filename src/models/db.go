package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Base) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}

type Invite struct {
	Base
	FirstName   string
	LastName    string
	Email       string
	State       string
	Country     string
	Affiliation string
	LoginNames  datatypes.JSON `json:"login_names" gorm:"type:json"`
}

type OTP struct {
	Base
	InviteID string
	Code     int
}

type EmailRate struct {
	Base
	Email    string
	LastSend time.Time
}
