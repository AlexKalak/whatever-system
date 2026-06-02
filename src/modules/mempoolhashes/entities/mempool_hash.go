package entities

import "time"

type MempoolHash struct {
	Hash      string    `json:"hash" gorm:"column:hash;type:varchar(66);primaryKey;index"`
	ChainID   uint      `json:"chainId" gorm:"column:chain_id;not null;index"`
	Timestamp time.Time `json:"timestamp" gorm:"column:timestamp;not null;index"`
}

func (MempoolHash) TableName() string {
	return "mempool_hashes"
}
