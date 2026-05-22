package models

import "time"

type SuratMasuk struct {
	IDSuratMasuk      uint       `gorm:"column:id_surat_masuk;primaryKey"`
	NoSurat           string     `gorm:"column:no_surat"`
	PerihalSurat      string     `gorm:"column:perihal_surat"`
	AsalSurat         string     `gorm:"column:asal_surat"`
	TanggalSurat      *time.Time `gorm:"column:tanggal_surat"`
	FilePDF           string     `gorm:"column:file_pdf"`
	TanggalDiterima   *time.Time `gorm:"column:tanggal_diterima"`
	StatusVerifikasi  string     `gorm:"column:status_verifikasi"`
	UserVerifikasi    *uint      `gorm:"column:user_verifikasi"`
	TanggalVerifikasi *time.Time `gorm:"column:tanggal_verifikasi"`
	CatatanVerifikasi string     `gorm:"column:catatan_verifikasi"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	IDDisposisiAktif  *uint      `gorm:"column:id_disposisi_aktif"`
	StatusAlur        string     `gorm:"column:status_alur"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (SuratMasuk) TableName() string {
	return "surat_masuk"
}