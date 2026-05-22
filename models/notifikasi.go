package models

import "time"

type Notifikasi struct {
	IDNotifikasi  uint       `gorm:"column:id_notifikasi;primaryKey"`
	IDPenerima    uint       `gorm:"column:id_penerima"`
	IDPengirim    uint       `gorm:"column:id_pengirim"`
	Jenis         string     `gorm:"column:jenis"`
	Judul         string     `gorm:"column:judul"`
	Pesan         string     `gorm:"column:pesan"`
	IsRead        bool       `gorm:"column:is_read"`
	WaktuBaca     *time.Time `gorm:"column:waktu_baca"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	LinkURL       string     `gorm:"column:link_url"`
	TipeReferensi string     `gorm:"column:tipe_referensi"`
}

func (Notifikasi) TableName() string {
	return "notifikasi"
}
