package repositories

import (
	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type DistribusiSMRepository interface {
	Create(d *models.DistribusiSM) error
	CreateBatch(distribusi []models.DistribusiSM) error
	GetBySuratMasukID(suratMasukID uint) ([]models.DistribusiSM, error)
	GetByJabatanID(jabatanID uint, page, pageSize int) ([]models.DistribusiSM, int64, error)
	GetByUserJabatanIDs(jabatanIDs []uint, page, pageSize int) ([]models.DistribusiSM, int64, error)
	MarkAsRead(id uint) error
	CountUnreadByJabatanIDs(jabatanIDs []uint) int64
}

type DistribusiSMRepositoryImpl struct {
	db *gorm.DB
}

func NewDistribusiSMRepository(db *gorm.DB) DistribusiSMRepository {
	return &DistribusiSMRepositoryImpl{db: db}
}

func (r *DistribusiSMRepositoryImpl) Create(d *models.DistribusiSM) error {
	return r.db.Create(d).Error
}

func (r *DistribusiSMRepositoryImpl) CreateBatch(distribusi []models.DistribusiSM) error {
	return r.db.Create(&distribusi).Error
}

func (r *DistribusiSMRepositoryImpl) GetBySuratMasukID(suratMasukID uint) ([]models.DistribusiSM, error) {
	var out []models.DistribusiSM
	err := r.db.Where("id_surat_masuk = ?", suratMasukID).
		Preload("Jabatan").
		Preload("SuratMasuk").
		Find(&out).Error
	return out, err
}

func (r *DistribusiSMRepositoryImpl) GetByJabatanID(jabatanID uint, page, pageSize int) ([]models.DistribusiSM, int64, error) {
	var out []models.DistribusiSM
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.Where("id_jabatan = ?", jabatanID)
	query.Model(&models.DistribusiSM{}).Count(&total)
	err := query.
		Preload("SuratMasuk").
		Preload("Jabatan").
		Order("distribute_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&out).Error
	return out, total, err
}

func (r *DistribusiSMRepositoryImpl) GetByUserJabatanIDs(jabatanIDs []uint, page, pageSize int) ([]models.DistribusiSM, int64, error) {
	var out []models.DistribusiSM
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.Where("id_jabatan IN ?", jabatanIDs)
	query.Model(&models.DistribusiSM{}).Count(&total)
	err := query.
		Preload("SuratMasuk").
		Preload("Jabatan").
		Order("distribute_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&out).Error
	return out, total, err
}

func (r *DistribusiSMRepositoryImpl) MarkAsRead(id uint) error {
	return r.db.Model(&models.DistribusiSM{}).
		Where("id_distribusi = ?", id).
		Update("status", "dibaca").Error
}

func (r *DistribusiSMRepositoryImpl) CountUnreadByJabatanIDs(jabatanIDs []uint) int64 {
	var count int64
	r.db.Model(&models.DistribusiSM{}).
		Where("id_jabatan IN ? AND status = ?", jabatanIDs, "belum_dibaca").
		Count(&count)
	return count
}
