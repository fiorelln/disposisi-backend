package models

import "time"

type Disposisi struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SuratID       uint      `json:"surat_id"`
	TujuanID      uint      `json:"tujuan_id"`
	VerifikatorID uint      `json:"verifikator_id"` // Kepala Sekolah yang approve
	Tujuan        string    `json:"tujuan"` // Nama/deskripsi tujuan
	Catatan       string    `json:"catatan"` // Catatan disposisi
	Status        string    `json:"status"` // menunggu, disetujui, ditolak, selesai
	VerifikasiStatus string `json:"verifikasi_status"` // menunggu, setuju, tolak
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Surat         Surat     `gorm:"foreignKey:SuratID" json:"surat"`
	TujuanUser    User      `gorm:"foreignKey:TujuanID" json:"tujuan_user"`
	Verifikator   User      `gorm:"foreignKey:VerifikatorID" json:"verifikator"`
}