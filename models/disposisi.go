package models

import "time"

type Disposisi struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	SuratID            uint      `json:"surat_id"`
	TujuanID           uint      `json:"tujuan_id"`
	VerifikatorID      uint      `json:"verifikator_id"`
	Tujuan             string    `json:"tujuan"`
	Catatan            string    `json:"catatan"`
	Status             string    `json:"status"`
	VerifikasiStatus   string    `json:"verifikasi_status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	Surat              *Surat    `gorm:"foreignKey:SuratID" json:"surat"`
	TujuanUser         User      `gorm:"foreignKey:TujuanID" json:"tujuan_user"`
	Verifikator        User      `gorm:"foreignKey:VerifikatorID" json:"verifikator"`
}
