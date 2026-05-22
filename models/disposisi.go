package models

import "time"

type Disposisi struct {
	IDDisposisi          uint       `gorm:"column:id_disposisi;primaryKey"`
	Sifat                string     `gorm:"column:sifat"`
	Catatan              string     `gorm:"column:catatan"`
	TanggapanSaran       string     `gorm:"column:tanggapan_saran"`
	ProsesLanjut         string     `gorm:"column:proses_lanjut"`
	KoordinasiKonfirmasi string     `gorm:"column:koordinasi_konfirmasi"`
	IDSuratMasuk         uint       `gorm:"column:id_surat_masuk"`
	IDKepsek             uint       `gorm:"column:id_kepsek"`
	IDPenerima           uint       `gorm:"column:id_penerima"`
	TanggalDisposisi     *time.Time `gorm:"column:tanggal_disposisi"`
	StatusDisposisi      string     `gorm:"column:status_disposisi"`
	StatusApproval       string     `gorm:"column:status_approval"`
	ApprovalAt           *time.Time `gorm:"column:approval_at"`

	SuratMasuk User `gorm:"foreignKey:IDSuratMasuk"`
	Penerima   User `gorm:"foreignKey:IDPenerima"`
	Kepsek     User `gorm:"foreignKey:IDKepsek"`
}

func (Disposisi) TableName() string {
	return "disposisi"
}
