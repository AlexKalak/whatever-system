package entities

import (
	"time"

	"github.com/google/uuid"
)

type AssetMapping struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	Type string `gorm:"type:varchar(20);not null;index"` // onchain | cex | another
	Name string `gorm:"type:varchar(100);not null;index"`

	//Cex specific
	Exchange         string `gorm:"type: varchar(50);index"`
	SymbolOnExchange string `gorm:"type: varchar(50);index"`
	//

	//Chain specific
	ChainID  uint   `gorm:"type:uint;index"`
	Address  string `gorm:"type:varchar(255);index"`
	Decimals int    `gorm:"type:uint;"`
	//

	CreatedAt time.Time
	UpdatedAt time.Time
}
