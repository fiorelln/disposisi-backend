package models

import "database/sql"
import "time"

// Status disposisi
const (
	StatusPending    = "pending"
	StatusForwarded  = "forwarded"
	StatusCompleted  = "completed"
	StatusRejected   = "rejected"
	StatusCancelled  = "cancelled"
)

type Disposisi struct {
	ID                   uint           `gorm:"column:id;primaryKey" json:"id"`
	SuratMasukID         uint           `gorm:"column:surat_masuk_id;index" json:"surat_masuk_id"`
	FromUserID           uint           `gorm:"column:from_user_id;index" json:"from_user_id"`
	ToUserID             uint           `gorm:"column:to_user_id;index" json:"to_user_id"`
	ParentDisposisiID    *uint          `gorm:"column:parent_disposisi_id;index" json:"parent_disposisi_id"`
	Level                int            `gorm:"column:level;default:0" json:"level"`
	Status               string         `gorm:"column:status;default:pending;index" json:"status"` // pending, forwarded, completed, rejected, cancelled
	Catatan              string         `gorm:"column:catatan;type:text" json:"catatan"`
	Dibaca               bool           `gorm:"column:dibaca;default:false" json:"dibaca"`
	Sifat                string         `gorm:"column:sifat"  json:"sifat"`                                                              // penting, biasa, rahasia, etc
	TanggapanSaran       string         `gorm:"column:tanggapan_saran;type:text" json:"tanggapan_saran"`
	ProsesLanjut         string         `gorm:"column:proses_lanjut;type:text" json:"proses_lanjut"`
	KoordinasiKonfirmasi string         `gorm:"column:koordinasi_konfirmasi;type:text" json:"koordinasi_konfirmasi"`
	BacaAt               *time.Time     `gorm:"column:baca_at" json:"baca_at"`
	CompleteAt           *time.Time     `gorm:"column:complete_at" json:"complete_at"`
	CreatedAt            time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt            sql.NullTime   `gorm:"column:deleted_at;index" json:"deleted_at"` // soft delete untuk audit

	// Relations
	SuratMasuk        *SuratMasuk `gorm:"foreignKey:SuratMasukID;references:ID" json:"surat_masuk,omitempty"`
	FromUser          *User       `gorm:"foreignKey:FromUserID;references:ID" json:"from_user,omitempty"`
	ToUser            *User       `gorm:"foreignKey:ToUserID;references:ID" json:"to_user,omitempty"`
	ParentDisposisi   *Disposisi  `gorm:"foreignKey:ParentDisposisiID;references:ID" json:"parent_disposisi,omitempty"`
	ChildDisposisi    []Disposisi `gorm:"foreignKey:ParentDisposisiID;references:ID" json:"child_disposisi,omitempty"`
}

func (Disposisi) TableName() string {
	return "disposisi"
}

func (d *Disposisi) IsCompleted() bool {
	return d.Status == StatusCompleted
}

func (d *Disposisi) IsForwarded() bool {
	return d.Status == StatusForwarded
}

func (d *Disposisi) IsPending() bool {
	return d.Status == StatusPending
}

func (d *Disposisi) IsRoot() bool {
	return d.ParentDisposisiID == nil && d.Level == 0
}
