package models

import "time"

type DistribusiSM struct {
	ID           uint      `gorm:"column:id_distribusi;primaryKey" json:"id"`
	IDSuratMasuk uint      `gorm:"column:id_surat_masuk" json:"id_surat_masuk"`
	IDJabatan    uint      `gorm:"column:id_jabatan" json:"id_jabatan"`
	Catatan      string    `gorm:"column:catatan;type:text" json:"catatan"`
	Status       string    `gorm:"column:status;default:belum_dibaca" json:"status"`
	DistributeAt time.Time `gorm:"column:distribute_at;default:CURRENT_TIMESTAMP" json:"distribute_at"`

	SuratMasuk *SuratMasuk `gorm:"foreignKey:IDSuratMasuk" json:"surat_masuk,omitempty"`
	Jabatan    *Jabatan    `gorm:"foreignKey:IDJabatan" json:"jabatan,omitempty"`
}

func (DistribusiSM) TableName() string {
	return "distribusi_sm"
}
