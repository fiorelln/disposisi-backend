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
	CanForwardToRole(fromRole string, toRole string) bool
}

type PermissionServiceImpl struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) PermissionService {
	return &PermissionServiceImpl{db: db}
}

type RoleLevel int

const (
	RoleHeadmaster  RoleLevel = 0
	RoleViceHead    RoleLevel = 1
	RoleCoordinator RoleLevel = 1
	RoleTU          RoleLevel = 2
	RoleTeacher     RoleLevel = 2
	RoleStaff       RoleLevel = 2
)

var roleLevelMap = map[string]RoleLevel{
	"KEPALA_SEKOLAH":  RoleHeadmaster,
	"WAKIL_KEPALA":    RoleViceHead,
	"WAKIL_KURIKULUM": RoleViceHead,
	"WAKIL_KESISWAAN": RoleViceHead,
	"WAKIL_SARANA":    RoleViceHead,
	"KORDINATOR":      RoleCoordinator,
	"TATA_USAHA":      RoleTU,
	"GURU":            RoleTeacher,
	"STAFF":           RoleStaff,
	"KEPALA_PERPUS":   RoleCoordinator,
	"KEPALA_LAB":      RoleCoordinator,
	"BK":              RoleCoordinator,
}

func (s *PermissionServiceImpl) CanForward(fromUserID uint, toUserID uint) (bool, error) {
	fromJabatans, err := s.GetUserJabatan(fromUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get from_user jabatan: %w", err)
	}

	toJabatans, err := s.GetUserJabatan(toUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get to_user jabatan: %w", err)
	}

	if len(fromJabatans) == 0 {
		return false, fmt.Errorf("from_user has no jabatan")
	}

	if len(toJabatans) == 0 {
		return false, fmt.Errorf("to_user has no jabatan")
	}

	fromRole := fromJabatans[0].NamaJabatan
	toRole := toJabatans[0].NamaJabatan

	return s.CanForwardToRole(fromRole, toRole), nil
}

func (s *PermissionServiceImpl) CanForwardToRole(fromRole string, toRole string) bool {
	if fromRole == "TATA_USAHA" || fromRole == "KEPALA_SEKOLAH" {
		return true
	}

	fromLevel, fromOk := roleLevelMap[fromRole]
	toLevel, toOk := roleLevelMap[toRole]

	if !fromOk || !toOk {
		return false
	}

	if toLevel < fromLevel {
		return true
	}

	if (fromRole == "WAKIL_KEPALA" ||
		fromRole == "WAKIL_KURIKULUM" ||
		fromRole == "WAKIL_KESISWAAN" ||
		fromRole == "WAKIL_SARANA") &&
		(toRole == "GURU" ||
			toRole == "STAFF" ||
			toRole == "TATA_USAHA" ||
			toRole == "WAKIL_KEPALA" ||
			toRole == "WAKIL_KURIKULUM" ||
			toRole == "WAKIL_KESISWAAN" ||
			toRole == "WAKIL_SARANA") {
		return true
	}

	if fromRole == "KORDINATOR" ||
		fromRole == "KEPALA_PERPUS" ||
		fromRole == "KEPALA_LAB" ||
		fromRole == "BK" {

		if toRole == "GURU" ||
			toRole == "STAFF" ||
			toRole == "TATA_USAHA" {
			return true
		}
	}

	if fromRole == "GURU" || fromRole == "STAFF" {
		if toRole == "KORDINATOR" ||
			toRole == "KEPALA_PERPUS" ||
			toRole == "KEPALA_LAB" ||
			toRole == "BK" ||
			toRole == "WAKIL_KEPALA" ||
			toRole == "WAKIL_KURIKULUM" ||
			toRole == "WAKIL_KESISWAAN" ||
			toRole == "WAKIL_SARANA" ||
			toRole == "KEPALA_SEKOLAH" {
			return true
		}
	}

	return false
}

func (s *PermissionServiceImpl) GetUserJabatan(userID uint) ([]models.Jabatan, error) {
	var userJabatans []models.UserJabatan
	var jabatans []models.Jabatan

	err := s.db.
		Where("id_user = ?", userID).
		Order("is_primary DESC, id_jabatan ASC").
		Find(&userJabatans).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user jabatan: %w", err)
	}

	jabatanIDs := make([]uint, len(userJabatans))

	for i, uj := range userJabatans {
		jabatanIDs[i] = uj.JabatanID
	}

	if len(jabatanIDs) == 0 {
		return jabatans, nil
	}

	err = s.db.
		Where("id_jabatan IN ?", jabatanIDs).
		Find(&jabatans).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get jabatan: %w", err)
	}

	return jabatans, nil
}

func (s *PermissionServiceImpl) GetPrimaryJabatan(userID uint) (*models.Jabatan, error) {
	var userJabatan models.UserJabatan

	err := s.db.
		Where("id_user = ? AND is_primary = ?", userID, true).
		First(&userJabatan).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get primary jabatan: %w", err)
	}

	var jabatan models.Jabatan

	err = s.db.
		Where("id_jabatan = ?", userJabatan.JabatanID).
		First(&jabatan).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get jabatan detail: %w", err)
	}

	return &jabatan, nil
}