package services

import (
	"fmt"

	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

// PermissionService interface - handle authorization logic
type PermissionService interface {
	// Forward permission
	CanForward(fromUserID uint, toUserID uint) (bool, error)

	// Get user jabatan
	GetUserJabatan(userID uint) ([]models.Jabatan, error)

	// Get user primary jabatan
	GetPrimaryJabatan(userID uint) (*models.Jabatan, error)

	// Check role hierarchy
	CanForwardToRole(fromRole string, toRole string) bool
}

// PermissionServiceImpl - service implementation
type PermissionServiceImpl struct {
	db *gorm.DB
}

// NewPermissionService - create new permission service
func NewPermissionService(db *gorm.DB) PermissionService {
	return &PermissionServiceImpl{db: db}
}

// Role Hierarchy
// Level 0 (Tertinggi): KEPALA_SEKOLAH
// Level 1: WAKIL_KEPALA, KORDINATOR
// Level 2: GURU, STAFF, TATA_USAHA
type RoleLevel int

const (
	RoleHeadmaster     RoleLevel = 0
	RoleViceHead       RoleLevel = 1
	RoleCoordinator    RoleLevel = 1
	RoleTU             RoleLevel = 2
	RoleTeacher        RoleLevel = 2
	RoleStaff          RoleLevel = 2
)

// role to level mapping
var roleLevelMap = map[string]RoleLevel{
	"KEPALA_SEKOLAH":    RoleHeadmaster,
	"WAKIL_KEPALA":      RoleViceHead,
	"WAKIL_KURIKULUM":   RoleViceHead,
	"WAKIL_KESISWAAN":   RoleViceHead,
	"WAKIL_SARANA":      RoleViceHead,
	"KORDINATOR":        RoleCoordinator,
	"TATA_USAHA":        RoleTU,
	"GURU":              RoleTeacher,
	"STAFF":             RoleStaff,
	"KEPALA_PERPUS":     RoleCoordinator,
	"KEPALA_LAB":        RoleCoordinator,
	"BK":                RoleCoordinator,
}

// ===== FORWARD PERMISSION =====

// CanForward - check apakah fromUser bisa forward ke toUser
// Rules:
// 1. TU bisa forward ke siapa saja (Kepsek, WAKA, Guru, Staff)
// 2. Kepsek bisa forward ke siapa saja
// 3. WAKA bisa forward ke WAKA lain, Guru, Staff di area mereka
// 4. Guru/Staff hanya bisa forward ke level atas atau spesifik koordinator
func (s *PermissionServiceImpl) CanForward(fromUserID uint, toUserID uint) (bool, error) {
	// Get jabatan masing-masing
	fromJabatans, err := s.GetUserJabatan(fromUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get from_user jabatan: %w", err)
	}

	toJabatans, err := s.GetUserJabatan(toUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get to_user jabatan: %w", err)
	}

	// Get primary jabatan untuk check
	if len(fromJabatans) == 0 {
		return false, fmt.Errorf("from_user has no jabatan")
	}

	if len(toJabatans) == 0 {
		return false, fmt.Errorf("to_user has no jabatan")
	}

	// Get primary jabatan (index 0 atau is_primary = true)
	fromRole := fromJabatans[0].NamaJabatan
	toRole := toJabatans[0].NamaJabatan

	// Check if forward is allowed
	return s.CanForwardToRole(fromRole, toRole), nil
}

// CanForwardToRole - check apakah role A bisa forward ke role B
func (s *PermissionServiceImpl) CanForwardToRole(fromRole string, toRole string) bool {
	// TU dan Kepsek bisa forward ke siapa saja
	if fromRole == "TATA_USAHA" || fromRole == "KEPALA_SEKOLAH" {
		return true
	}

	// Get level dari role
	fromLevel, fromOk := roleLevelMap[fromRole]
	toLevel, toOk := roleLevelMap[toRole]

	if !fromOk || !toOk {
		return false
	}

	// Rules untuk forward:
	// 1. Bisa forward ke level atas (lebih kecil level number)
	if toLevel < fromLevel {
		return true
	}

	// 2. WAKA bisa forward ke WAKA lain dan ke level bawah
	if (fromRole == "WAKIL_KEPALA" || fromRole == "WAKIL_KURIKULUM" || fromRole == "WAKIL_KESISWAAN" || fromRole == "WAKIL_SARANA") &&
		(toRole == "GURU" || toRole == "STAFF" || toRole == "TATA_USAHA" ||
			toRole == "WAKIL_KEPALA" || toRole == "WAKIL_KURIKULUM" || toRole == "WAKIL_KESISWAAN" || toRole == "WAKIL_SARANA") {
		return true
	}

	// 3. Koordinator bisa forward ke bawahan mereka (Guru, Staff)
	if fromRole == "KORDINATOR" || fromRole == "KEPALA_PERPUS" || fromRole == "KEPALA_LAB" || fromRole == "BK" {
		if toRole == "GURU" || toRole == "STAFF" || toRole == "TATA_USAHA" {
			return true
		}
	}

	// 4. Guru/Staff tidak bisa forward kecuali ke level atas
	if fromRole == "GURU" || fromRole == "STAFF" {
		// Hanya bisa forward ke Koordinator, WAKA, atau Kepsek
		if toRole == "KORDINATOR" || toRole == "KEPALA_PERPUS" || toRole == "KEPALA_LAB" || toRole == "BK" ||
			toRole == "WAKIL_KEPALA" || toRole == "WAKIL_KURIKULUM" || toRole == "WAKIL_KESISWAAN" || toRole == "WAKIL_SARANA" ||
			toRole == "KEPALA_SEKOLAH" {
			return true
		}
	}

	return false
}

// ===== JABATAN OPERATIONS =====

// GetUserJabatan - get semua jabatan untuk user
func (s *PermissionServiceImpl) GetUserJabatan(userID uint) ([]models.Jabatan, error) {
	var userJabatans []models.UserJabatan
	var jabatans []models.Jabatan

	// Get user jabatan dengan primary first
	err := s.db.
		Where("id_user = ?", userID).
		Order("is_primary DESC, id_jabatan ASC").
		Find(&userJabatans).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user jabatan: %w", err)
	}

	// Extract jabatan
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

// GetPrimaryJabatan - get primary jabatan untuk user
func (s *PermissionServiceImpl) GetPrimaryJabatan(userID uint) (*models.Jabatan, error) {
	var userJabatan models.UserJabatan

	// Get primary jabatan
	err := s.db.
		Where("id_user = ? AND is_primary = ?", userID, true).
		First(&userJabatan).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get primary jabatan: %w", err)
	}

	// Get jabatan detail
	var jabatan models.Jabatan
	err = s.db.
		Where("id_jabatan = ?", userJabatan.JabatanID).
		First(&jabatan).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get jabatan detail: %w", err)
	}

	return &jabatan, nil
}
