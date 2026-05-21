package entities

import (
	"time"

	"github.com/google/uuid"
)

type Asset struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name   string    `json:"name" gorm:"type:varchar(50)"`
	Symbol string    `json:"symbol" gorm:"type:varchar(50);index"`
	Type   string    `json:"type" gorm:"type:varchar(20);index"` // crypto | stock | fx | derivative

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
