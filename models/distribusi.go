package models

import "time"

type DistribusiSK struct {
	IDDistribusi uint       `gorm:"column:id_distribusi;primaryKey"`
	IDSK         uint       `gorm:"column:id_sk"`
	IDUser       uint       `gorm:"column:id_user"`
	Status       string     `gorm:"column:status"`
	DistributeAt *time.Time `gorm:"column:distribute_at"`
	ReadAt       *time.Time `gorm:"column:read_at"`
	Catatan      string     `gorm:"column:catatan"`
}

func (DistribusiSK) TableName() string {
	return "distribusi_sk"
}
