package entities

import "time"

type Token struct {
	ChainID  uint   `json:"chainId" gorm:"not null;uniqueIndex:idx_chain_token_address"`
	Address  string `json:"address" gorm:"type:varchar(255);not null;uniqueIndex:idx_chain_token_address"`
	Symbol   string `json:"symbol" gorm:"type:varchar(50);not null;index"`
	Name     string `json:"name" gorm:"type:varchar(100)"`
	Decimals uint8  `json:"decimals" gorm:"type:smallint;not null"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
