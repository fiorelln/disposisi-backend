package services

import (
	"fmt"

	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type PermissionService interface {
	CanForward(fromUserID uint, toUserID uint) (bool, error)
	GetUserJabatan(userID uint) ([]models.Jabatan, error)
	GetPrimaryJabatan(userID uint) (*models.Jabatan, error)
}

type PermissionServiceImpl struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) PermissionService {
	return &PermissionServiceImpl{db: db}
}

func (s *PermissionServiceImpl) CanForward(fromUserID uint, toUserID uint) (bool, error) {
	fromJabatans, err := s.GetUserJabatan(fromUserID)
	if err != nil {
		return false, fmt.Errorf("gagal mengambil jabatan pengirim: %w", err)
	}

	if len(fromJabatans) == 0 {
		return false, fmt.Errorf("pengirim tidak memiliki jabatan")
	}

	fromRole := fromJabatans[0].NamaJabatan
	if fromRole == "Tata Usaha" {
		return true, nil
	}

	return false, nil
}

func (s *PermissionServiceImpl) GetUserJabatan(userID uint) ([]models.Jabatan, error) {
	var userJabatans []models.UserJabatan

	err := s.db.
		Where("id_user = ?", userID).
		Order("is_primary DESC, id_jabatan ASC").
		Find(&userJabatans).Error

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data jabatan user: %w", err)
	}

	if len(userJabatans) == 0 {
		return []models.Jabatan{}, nil
	}

	jabatanIDs := make([]uint, len(userJabatans))
	for i, uj := range userJabatans {
		jabatanIDs[i] = uj.JabatanID
	}

	var jabatans []models.Jabatan
	err = s.db.
		Where("id_jabatan IN ?", jabatanIDs).
		Find(&jabatans).Error

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail jabatan: %w", err)
	}

	return jabatans, nil
}

func (s *PermissionServiceImpl) GetPrimaryJabatan(userID uint) (*models.Jabatan, error) {
	var userJabatan models.UserJabatan

	err := s.db.
		Where("id_user = ? AND is_primary = ?", userID, true).
		First(&userJabatan).Error

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil jabatan utama user: %w", err)
	}

	var jabatan models.Jabatan

	err = s.db.
		Where("id_jabatan = ?", userJabatan.JabatanID).
		First(&jabatan).Error

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail jabatan: %w", err)
	}

	return &jabatan, nil
}
