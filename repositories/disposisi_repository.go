package repositories

import (
	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type DisposisiRepository interface {
	Create(disposisi *models.Disposisi) error
	GetByID(id uint) (*models.Disposisi, error)
	Update(disposisi *models.Disposisi) error
	Delete(id uint) error
}

type DisposisiRepositoryImpl struct {
	db *gorm.DB
}

func NewDisposisiRepository(db *gorm.DB) DisposisiRepository {
	return &DisposisiRepositoryImpl{db: db}
}

func (r *DisposisiRepositoryImpl) Create(disposisi *models.Disposisi) error {
	return r.db.Create(disposisi).Error
}

func (r *DisposisiRepositoryImpl) GetByID(id uint) (*models.Disposisi, error) {
	var disposisi models.Disposisi
	err := r.db.
		Preload("SuratMasuk").
		Preload("Kepsek").
		Preload("Penerima").
		Preload("JabatanPenerima").
		First(&disposisi, id).Error
	if err != nil {
		return nil, err
	}
	return &disposisi, nil
}

func (r *DisposisiRepositoryImpl) Update(disposisi *models.Disposisi) error {
	return r.db.Where("id_disposisi = ?", disposisi.ID).Save(disposisi).Error
}

func (r *DisposisiRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.Disposisi{}, id).Error
}
