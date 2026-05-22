package models

import "time"

type SuratKeluar struct {
	IDSuratKeluar     uint       `gorm:"column:id_surat_keluar;primaryKey"`
	KodeSurat         string     `gorm:"column:kode_surat"`
	NoSurat           string     `gorm:"column:no_surat"`
	Perihal           string     `gorm:"column:perihal"`
	Catatan           string     `gorm:"column:catatan"`
	TanggalSurat      *time.Time `gorm:"column:tanggal_surat"`
	FilePDF           string     `gorm:"column:file_pdf"`
	StatusVerifikasi  string     `gorm:"column:status_verifikasi"`
	UserVerifikasi    *uint      `gorm:"column:user_verifikasi"`
	TanggalVerifikasi *time.Time `gorm:"column:tanggal_verifikasi"`
	Tujuan            string     `gorm:"column:tujuan"`
	CatatanVerifikasi string     `gorm:"column:catatan_verifikasi"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
	StatusAlur        string     `gorm:"column:status_alur"`
}

func (SuratKeluar) TableName() string {
	return "surat_keluar"
}
