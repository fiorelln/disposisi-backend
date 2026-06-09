package repositories

import (
	"time"

	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type SuratMasukRepository interface {
	Create(surat *models.SuratMasuk) error
	GetByID(id uint) (*models.SuratMasuk, error)
	Update(surat *models.SuratMasuk) error
	Delete(id uint) error
	ListAll(page, pageSize int, status string) ([]models.SuratMasuk, int64, error)
	UpdateStatusAlur(id uint, status string) error
	SetDisposisiAktif(id uint, disposisiID uint) error
	Verify(id uint, userID uint, status string, notes string) error
}

type suratMasukRepository struct {
	db *gorm.DB
}

func NewSuratMasukRepository(db *gorm.DB) SuratMasukRepository {
	return &suratMasukRepository{db: db}
}

func (r *suratMasukRepository) Create(surat *models.SuratMasuk) error {
	return r.db.Create(surat).Error
}

func (r *suratMasukRepository) GetByID(id uint) (*models.SuratMasuk, error) {
	var surat models.SuratMasuk
	err := r.db.Preload("DisposisiAktif").Where("id_surat_masuk = ?", id).First(&surat).Error
	if err != nil {
		return nil, err
	}
	return &surat, nil
}

func (r *suratMasukRepository) Update(surat *models.SuratMasuk) error {
	return r.db.Where("id_surat_masuk = ?", surat.IDSuratMasuk).Save(surat).Error
}

func (r *suratMasukRepository) Delete(id uint) error {
	return r.db.Delete(&models.SuratMasuk{}, id).Error
}

func (r *suratMasukRepository) ListAll(page, pageSize int, status string) ([]models.SuratMasuk, int64, error) {
	var surat []models.SuratMasuk
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.Model(&models.SuratMasuk{})
	if status != "" {
		query = query.Where("status_alur = ?", status)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&surat).Error

	return surat, total, err
}

func (r *suratMasukRepository) UpdateStatusAlur(id uint, status string) error {
	return r.db.Model(&models.SuratMasuk{}).Where("id_surat_masuk = ?", id).Update("status_alur", status).Error
}

func (r *suratMasukRepository) SetDisposisiAktif(id uint, disposisiID uint) error {
	return r.db.Model(&models.SuratMasuk{}).Where("id_surat_masuk = ?", id).Update("id_disposisi_aktif", disposisiID).Error
}

func (r *suratMasukRepository) Verify(id uint, userID uint, status string, notes string) error {
	now := time.Now()
	return r.db.Model(&models.SuratMasuk{}).Where("id_surat_masuk = ?", id).Updates(map[string]interface{}{
		"user_verifikasi":    userID,
		"status_verifikasi":  status,
		"catatan_verifikasi": notes,
		"tanggal_verifikasi": &now,
	}).Error
}
