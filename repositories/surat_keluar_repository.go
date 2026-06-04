package repositories

import (
	"time"

	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type SuratKeluarRepository interface {
	Create(surat *models.SuratKeluar) error
	GetByID(id uint) (*models.SuratKeluar, error)
	Update(surat *models.SuratKeluar) error
	Delete(id uint) error
	ListAll(page, pageSize int, status string) ([]models.SuratKeluar, int64, error)
	Verify(id uint, userID uint, status string, notes string) error
	UpdateStatusAlur(id uint, status string) error
}

type suratKeluarRepository struct {
	db *gorm.DB
}

func NewSuratKeluarRepository(db *gorm.DB) SuratKeluarRepository {
	return &suratKeluarRepository{db: db}
}

func (r *suratKeluarRepository) Create(surat *models.SuratKeluar) error {
	return r.db.Create(surat).Error
}

func (r *suratKeluarRepository) GetByID(id uint) (*models.SuratKeluar, error) {
	var surat models.SuratKeluar
	err := r.db.Where("id_surat_keluar = ?", id).First(&surat).Error
	if err != nil {
		return nil, err
	}
	return &surat, nil
}

func (r *suratKeluarRepository) Update(surat *models.SuratKeluar) error {
	return r.db.Where("id_surat_keluar = ?", surat.IDSuratKeluar).Save(surat).Error
}

func (r *suratKeluarRepository) Delete(id uint) error {
	return r.db.Delete(&models.SuratKeluar{}, id).Error
}

func (r *suratKeluarRepository) ListAll(page, pageSize int, status string) ([]models.SuratKeluar, int64, error) {
	var surat []models.SuratKeluar
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.Model(&models.SuratKeluar{})
	if status != "" {
		query = query.Where("status_alur = ?", status)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&surat).Error

	return surat, total, err
}

func (r *suratKeluarRepository) Verify(id uint, userID uint, status string, notes string) error {
	now := time.Now()
	return r.db.Model(&models.SuratKeluar{}).Where("id_surat_keluar = ?", id).Updates(map[string]interface{}{
		"user_verifikasi":    userID,
		"status_verifikasi":  status,
		"catatan_verifikasi": notes,
		"tanggal_verifikasi": &now,
	}).Error
}

func (r *suratKeluarRepository) UpdateStatusAlur(id uint, status string) error {
	return r.db.Model(&models.SuratKeluar{}).Where("id_surat_keluar = ?", id).Update("status_alur", status).Error
}
