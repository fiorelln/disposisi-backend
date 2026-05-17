package models

import (
	"time"
)

type Surat struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	FileSurat string     `json:"file_surat"`
	Status    string     `json:"status"` // dikirim, diteruskan, disetujui, ditolak, selesai
	TujuanID  uint       `json:"tujuan_id"` // user ID yang dituju
	PengirimID uint      `json:"pengirim_id"` // user ID pengirim
	Kategori  string     `json:"kategori"` // kategori surat
	Judul     string     `json:"judul"`
	Deskripsi string     `json:"deskripsi"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	Pengirim   User       `gorm:"foreignKey:PengirimID" json:"pengirim"`
	Tujuan     User       `gorm:"foreignKey:TujuanID" json:"tujuan"`
	Disposisi  Disposisi  `gorm:"foreignKey:SuratID" json:"disposisi"`
}
