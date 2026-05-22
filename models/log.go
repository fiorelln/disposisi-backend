package models

import "time"

type Log struct {
	IDLog        uint      `gorm:"column:id_log;primaryKey"`
	IDUser       uint      `gorm:"column:id_user"`
	Aksi         string    `gorm:"column:aksi"`
	TabelTerkait string    `gorm:"column:tabel_terkait"`
	KolomTerkait string    `gorm:"column:kolom_terkait"`
	IDData       uint      `gorm:"column:id_data"`
	ValuesOld    string    `gorm:"column:values_old"`
	ValuesNew    string    `gorm:"column:values_new"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Log) TableName() string {
	return "log"
}
