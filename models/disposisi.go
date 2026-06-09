package models

import "time"

type Disposisi struct {
	ID                   uint       `gorm:"column:id_disposisi;primaryKey" json:"id"`
	TanggapanSaran       string     `gorm:"column:tanggapan_saran;type:text" json:"tanggapan_saran"`
	ProsesLanjut         string     `gorm:"column:proses_lanjut;type:text" json:"proses_lanjut"`
	KoordinasiKonfirmasi string     `gorm:"column:koordinasi_konfirmasi;type:text" json:"koordinasi_konfirmasi"`
	SuratMasukID         uint       `gorm:"column:id_surat_masuk" json:"id_surat_masuk"`
	KepsekID             *uint      `gorm:"column:id_kepsek" json:"id_kepsek"`
	PenerimaID           *uint      `gorm:"column:id_penerima" json:"id_penerima"`
	TanggalDisposisi     time.Time  `gorm:"column:tanggal_disposisi;autoCreateTime" json:"tanggal_disposisi"`
	StatusDisposisi      string     `gorm:"column:status_disposisi;default:belum_dibaca" json:"status_disposisi"`
	StatusApproval       string     `gorm:"column:status_approval;default:menunggu" json:"status_approval"`
	ApprovalAt           *time.Time `gorm:"column:approval_at" json:"approval_at"`
	CatatanKepsek        string     `gorm:"column:catatan_kepsek;type:text" json:"catatan_kepsek"`
	JabatanPenerimaID    *uint      `gorm:"column:id_jabatan_penerima" json:"id_jabatan_penerima"`
	IsiDisposisi         string     `gorm:"column:isi_disposisi;type:text;default:''" json:"isi_disposisi"`
	BatasWaktu           string     `gorm:"column:batas_waktu;default:''" json:"batas_waktu"`

	SuratMasuk      *SuratMasuk `gorm:"foreignKey:SuratMasukID;references:IDSuratMasuk" json:"surat_masuk,omitempty"`
	Kepsek          *User       `gorm:"foreignKey:KepsekID;references:ID" json:"kepsek,omitempty"`
	Penerima        *User       `gorm:"foreignKey:PenerimaID;references:ID" json:"penerima,omitempty"`
	JabatanPenerima *Jabatan    `gorm:"foreignKey:JabatanPenerimaID;references:ID" json:"jabatan_penerima,omitempty"`
}

func (Disposisi) TableName() string {
	return "disposisi"
}
