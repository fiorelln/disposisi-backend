package models

import (
	"time"
)

type Surat struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	FileSurat string     `json:"file_surat"`
	Status    string     `json:"status"` 
	TujuanID  uint       `json:"tujuan_id"` 
	PengirimID uint      `json:"pengirim_id"` 
	Kategori  string     `json:"kategori"` 
	Judul     string     `json:"judul"`
	Deskripsi string     `json:"deskripsi"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	Pengirim   User       `gorm:"foreignKey:PengirimID" json:"pengirim"`
	Tujuan     User       `gorm:"foreignKey:TujuanID" json:"tujuan"`
	Disposisi  Disposisi  `gorm:"foreignKey:SuratID" json:"disposisi"`
}
