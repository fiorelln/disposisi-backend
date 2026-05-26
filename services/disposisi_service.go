package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/fiorelln/disposisi/dto"
	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/repositories"
	"gorm.io/gorm"
)

// DisposisiService interface - define service methods
type DisposisiService interface {
	// Core Operations
	CreateInitialDisposisi(suratMasukID uint, toUserID uint, notes string) (*models.Disposisi, error)
	ForwardDisposisi(disposisiID uint, req *dto.CreateForwardRequest, currentUserID uint) (*models.Disposisi, error)
	CompleteDisposisi(disposisiID uint, req *dto.CompleteDisposisiRequest, currentUserID uint) error
	RejectDisposisi(disposisiID uint, reason string, currentUserID uint) error

	// Status Operations
	MarkAsRead(disposisiID uint) error
	MarkAsReadBatch(disposisiIDs []uint) error

	// Query Operations
	GetInbox(userID uint, page, pageSize int) (*dto.InboxListResponse, error)
	GetSentItems(userID uint, page, pageSize int) (*dto.SentListResponse, error)
	GetHistory(suratMasukID uint) (*dto.HistoryResponse, error)
	GetDisposisiDetail(disposisiID uint) (*dto.DisposisiResponse, error)

	// Validation
	ValidateForward(disposisiID uint, toUserID uint, currentUserID uint) (bool, string, error)
	CheckCanForward(currentUserID uint, toUserID uint) (bool, error)

	// Dashboard
	GetStats(userID uint) (map[string]interface{}, error)
}

// DisposisiServiceImpl - service implementation
type DisposisiServiceImpl struct {
	repo            repositories.DisposisiRepository
	permissionSvc   PermissionService
	notificationSvc NotificationService
	db              *gorm.DB
}

// NewDisposisiService - create new service instance
func NewDisposisiService(
	repo repositories.DisposisiRepository,
	permissionSvc PermissionService,
	notificationSvc NotificationService,
	db *gorm.DB,
) DisposisiService {
	return &DisposisiServiceImpl{
		repo:            repo,
		permissionSvc:   permissionSvc,
		notificationSvc: notificationSvc,
		db:              db,
	}
}

// ===== CORE OPERATIONS =====

// CreateInitialDisposisi - create root disposisi dari surat masuk
// Biasanya dari TU ke Kepsek
func (s *DisposisiServiceImpl) CreateInitialDisposisi(
	suratMasukID uint,
	toUserID uint,
	notes string,
) (*models.Disposisi, error) {
	// Validate surat exists
	var surat models.SuratMasuk
	if err := s.db.Where("id_surat_masuk = ?", suratMasukID).First(&surat).Error; err != nil {
		return nil, fmt.Errorf("surat not found: %w", err)
	}

	// Create root disposisi (level 0)
	disposisi := &models.Disposisi{
		SuratMasukID:      suratMasukID,
		FromUserID:        1, // Assume TU (ID 1), atau bisa dari context
		ToUserID:          toUserID,
		ParentDisposisiID: nil,
		Level:             0,
		Status:            models.StatusPending,
		Catatan:           notes,
		Dibaca:            false,
	}

	// Save ke database
	if err := s.repo.Create(disposisi); err != nil {
		return nil, fmt.Errorf("failed to create disposisi: %w", err)
	}

	// Create notification untuk penerima
	s.notificationSvc.NotifyDisposisiReceived(disposisi.ID, toUserID)

	return disposisi, nil
}

// ForwardDisposisi - forward disposisi ke user lain (create child disposisi)
func (s *DisposisiServiceImpl) ForwardDisposisi(
	disposisiID uint,
	req *dto.CreateForwardRequest,
	currentUserID uint,
) (*models.Disposisi, error) {
	// Validate request
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Get parent disposisi
	parentDisposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return nil, fmt.Errorf("parent disposisi not found: %w", err)
	}

	// Check if current user is the owner of parent disposisi
	if parentDisposisi.ToUserID != currentUserID {
		return nil, errors.New("unauthorized: you are not the recipient of this disposisi")
	}

	// Validate permission to forward
	canForward, err := s.CheckCanForward(currentUserID, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !canForward {
		return nil, errors.New("you don't have permission to forward to this user")
	}

	// Create child disposisi
	childDisposisi := &models.Disposisi{
		SuratMasukID:      parentDisposisi.SuratMasukID,
		FromUserID:        currentUserID,
		ToUserID:          req.ToUserID,
		ParentDisposisiID: &disposisiID,
		Level:             parentDisposisi.Level + 1,
		Status:            models.StatusPending,
		Catatan:           req.Catatan,
		Dibaca:            false,
		Sifat:             req.Sifat,
	}

	// Use transaction untuk consistency
	tx := s.db.Begin()

	// Create child disposisi
	if err := s.repo.Create(childDisposisi); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create child disposisi: %w", err)
	}

	// Update parent status ke forwarded
	if err := s.repo.UpdateStatus(disposisiID, models.StatusForwarded); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update parent status: %w", err)
	}

	// Mark parent as read if not yet
	if !parentDisposisi.Dibaca {
		s.repo.MarkAsRead(disposisiID)
	}

	tx.Commit()

	// Create notification untuk recipient
	s.notificationSvc.NotifyDisposisiReceived(childDisposisi.ID, req.ToUserID)

	return childDisposisi, nil
}

// CompleteDisposisi - mark disposisi as completed
func (s *DisposisiServiceImpl) CompleteDisposisi(
	disposisiID uint,
	req *dto.CompleteDisposisiRequest,
	currentUserID uint,
) error {
	// Get disposisi
	disposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return fmt.Errorf("disposisi not found: %w", err)
	}

	// Check if current user is the recipient
	if disposisi.ToUserID != currentUserID {
		return errors.New("unauthorized: you are not the recipient of this disposisi")
	}

	// Check if already completed
	if disposisi.IsCompleted() {
		return errors.New("disposisi is already completed")
	}

	// Prepare update data
	updateData := map[string]interface{}{
		"status":                models.StatusCompleted,
		"complete_at":           time.Now(),
		"tanggapan_saran":       req.TanggapanSaran,
		"proses_lanjut":         req.ProsesLanjut,
		"koordinasi_konfirmasi": req.KoordinasiKonfirmasi,
		"catatan":               req.Catatan,
	}

	// Use transaction
	tx := s.db.Begin()

	// Update disposisi
	if err := s.repo.MarkAsCompleted(disposisiID, updateData); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to complete disposisi: %w", err)
	}

	// Check if all siblings are completed
	if disposisi.ParentDisposisiID != nil {
		s.autoUpdateParentStatus(*disposisi.ParentDisposisiID)
	}

	tx.Commit()

	// Create notification ke sender
	s.notificationSvc.NotifyDisposisiCompleted(disposisiID, disposisi.FromUserID)

	return nil
}

// RejectDisposisi - reject disposisi
func (s *DisposisiServiceImpl) RejectDisposisi(
	disposisiID uint,
	reason string,
	currentUserID uint,
) error {
	// Get disposisi
	disposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return fmt.Errorf("disposisi not found: %w", err)
	}

	// Check if current user is the recipient
	if disposisi.ToUserID != currentUserID {
		return errors.New("unauthorized: you are not the recipient of this disposisi")
	}

	// Update disposisi status to rejected
	if err := s.repo.Update(&models.Disposisi{
		ID:         disposisiID,
		Status:     models.StatusRejected,
		Catatan:    reason,
		CompleteAt: &time.Time{},
	}); err != nil {
		return fmt.Errorf("failed to reject disposisi: %w", err)
	}

	// Create notification ke sender
	s.notificationSvc.NotifyDisposisiRejected(disposisiID, disposisi.FromUserID)

	return nil
}

// ===== STATUS OPERATIONS =====

// MarkAsRead - mark disposisi as read
func (s *DisposisiServiceImpl) MarkAsRead(disposisiID uint) error {
	return s.repo.MarkAsRead(disposisiID)
}

// MarkAsReadBatch - mark multiple disposisi as read
func (s *DisposisiServiceImpl) MarkAsReadBatch(disposisiIDs []uint) error {
	for _, id := range disposisiIDs {
		if err := s.repo.MarkAsRead(id); err != nil {
			return err
		}
	}
	return nil
}

// ===== QUERY OPERATIONS =====

// GetInbox - get inbox dengan pagination dan converted to DTO
func (s *DisposisiServiceImpl) GetInbox(
	userID uint,
	page, pageSize int,
) (*dto.InboxListResponse, error) {
	// Validasi pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get dari repository
	disposisi, total, err := s.repo.GetInbox(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}

	// Convert ke DTO
	items := make([]dto.InboxItemResponse, len(disposisi))
	for i, d := range disposisi {
		items[i] = dto.InboxItemResponse{
			ID:           d.ID,
			SuratMasukID: d.SuratMasukID,
			FromUser: dto.UserBasicInfo{
				ID:    d.FromUser.ID,
				Name:  d.FromUser.Name,
				Email: d.FromUser.Email,
			},
			Status:       d.Status,
			Catatan:      d.Catatan,
			Dibaca:       d.Dibaca,
			Sifat:        d.Sifat,
			CreatedAt:    d.CreatedAt,
			UpdatedAt:    d.UpdatedAt,
			SuratNomor:   d.SuratMasuk.NoSurat,
			SuratPerihal: d.SuratMasuk.PerihalSurat,
		}
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.InboxListResponse{
		Data: items,
		Pagination: dto.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// GetSentItems - get sent items dengan pagination
func (s *DisposisiServiceImpl) GetSentItems(
	userID uint,
	page, pageSize int,
) (*dto.SentListResponse, error) {
	// Validasi pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get dari repository
	disposisi, total, err := s.repo.GetSentItems(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent items: %w", err)
	}

	// Convert ke DTO
	items := make([]dto.SentItemResponse, len(disposisi))
	for i, d := range disposisi {
		// Count child dan completed
		children, _ := s.repo.GetChildDisposisi(d.ID)
		completed, _ := s.repo.CountCompletedChildren(d.ID)

		items[i] = dto.SentItemResponse{
			ID:           d.ID,
			SuratMasukID: d.SuratMasukID,
			ToUser: dto.UserBasicInfo{
				ID:    d.ToUser.ID,
				Name:  d.ToUser.Name,
				Email: d.ToUser.Email,
			},
			Status:         d.Status,
			Catatan:        d.Catatan,
			CreatedAt:      d.CreatedAt,
			UpdatedAt:      d.UpdatedAt,
			ChildCount:     len(children),
			CompletedCount: completed,
		}
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.SentListResponse{
		Data: items,
		Pagination: dto.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// GetHistory - get full history/tree untuk satu surat
func (s *DisposisiServiceImpl) GetHistory(suratMasukID uint) (*dto.HistoryResponse, error) {
	// Get surat
	var surat models.SuratMasuk
	if err := s.db.Where("id_surat_masuk = ?", suratMasukID).First(&surat).Error; err != nil {
		return nil, fmt.Errorf("surat not found: %w", err)
	}

	// Get root disposisi
	rootDisposisi, err := s.repo.GetRootDisposisi(suratMasukID)
	if err != nil {
		return nil, fmt.Errorf("root disposisi not found: %w", err)
	}

	// Get full tree
	treeDisposisi, err := s.repo.GetDisposisiTree(rootDisposisi.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// Get all disposisi untuk count
	allDisposisi, err := s.repo.GetHistory(suratMasukID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	// Convert tree to DTO
	treeDTO := s.convertDisposisiToTreeNode(treeDisposisi)

	return &dto.HistoryResponse{
		SuratMasukID:  suratMasukID,
		SuratNomor:    surat.NoSurat,
		SuratPerihal:  surat.PerihalSurat,
		TanggalSurat:  surat.TanggalSurat,
		AsalSurat:     surat.AsalSurat,
		RootDisposisi: treeDTO,
		TotalForward:  len(allDisposisi),
		Status:        rootDisposisi.Status,
	}, nil
}

// GetDisposisiDetail - get detail disposisi
func (s *DisposisiServiceImpl) GetDisposisiDetail(disposisiID uint) (*dto.DisposisiResponse, error) {
	disposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return nil, fmt.Errorf("disposisi not found: %w", err)
	}

	children, _ := s.repo.GetChildDisposisi(disposisiID)

	return &dto.DisposisiResponse{
		ID:           disposisi.ID,
		SuratMasukID: disposisi.SuratMasukID,
		FromUser: dto.UserBasicInfo{
			ID:    disposisi.FromUser.ID,
			Name:  disposisi.FromUser.Name,
			Email: disposisi.FromUser.Email,
		},
		ToUser: dto.UserBasicInfo{
			ID:    disposisi.ToUser.ID,
			Name:  disposisi.ToUser.Name,
			Email: disposisi.ToUser.Email,
		},
		ParentDisposisiID:    disposisi.ParentDisposisiID,
		Level:                disposisi.Level,
		Status:               disposisi.Status,
		Catatan:              disposisi.Catatan,
		Dibaca:               disposisi.Dibaca,
		Sifat:                disposisi.Sifat,
		TanggapanSaran:       disposisi.TanggapanSaran,
		ProsesLanjut:         disposisi.ProsesLanjut,
		KoordinasiKonfirmasi: disposisi.KoordinasiKonfirmasi,
		BacaAt:               disposisi.BacaAt,
		CompleteAt:           disposisi.CompleteAt,
		CreatedAt:            disposisi.CreatedAt,
		UpdatedAt:            disposisi.UpdatedAt,
		ChildCount:           len(children),
	}, nil
}

// ===== VALIDATION OPERATIONS =====

// ValidateForward - validate apakah bisa forward ke user
func (s *DisposisiServiceImpl) ValidateForward(
	disposisiID uint,
	toUserID uint,
	currentUserID uint,
) (bool, string, error) {
	// Get disposisi
	disposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return false, "Disposisi tidak ditemukan", err
	}

	// Check if current user is recipient
	if disposisi.ToUserID != currentUserID {
		return false, "Anda bukan penerima disposisi ini", nil
	}

	// Check if disposisi is still pending
	if !disposisi.IsPending() {
		return false, fmt.Sprintf("Disposisi sudah dalam status %s", disposisi.Status), nil
	}

	// Check permission
	canForward, err := s.CheckCanForward(currentUserID, toUserID)
	if err != nil {
		return false, "Error checking permission", err
	}

	if !canForward {
		return false, "Anda tidak memiliki izin untuk forward ke user tersebut", nil
	}

	return true, "", nil
}

// CheckCanForward - check apakah user A bisa forward ke user B
func (s *DisposisiServiceImpl) CheckCanForward(currentUserID uint, toUserID uint) (bool, error) {
	// Delegate to PermissionService
	return s.permissionSvc.CanForward(currentUserID, toUserID)
}

// ===== DASHBOARD =====

// GetStats - get stats untuk dashboard
func (s *DisposisiServiceImpl) GetStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count unread inbox
	unread, err := s.repo.CountUnreadInbox(userID)
	if err != nil {
		return nil, fmt.Errorf("error counting unread: %w", err)
	}
	stats["unread_inbox"] = unread

	// Count pending inbox
	pending, err := s.repo.CountPendingInbox(userID)
	if err != nil {
		return nil, fmt.Errorf("error counting pending: %w", err)
	}
	stats["pending_inbox"] = pending

	// Count total inbox
	inbox, _, err := s.repo.GetInbox(userID, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("error getting inbox: %w", err)
	}
	stats["total_inbox"] = len(inbox)

	return stats, nil
}

// ===== HELPER METHODS =====

// autoUpdateParentStatus - auto update parent status jika semua siblings completed
func (s *DisposisiServiceImpl) autoUpdateParentStatus(parentID uint) {
	// Get all children
	children, err := s.repo.GetChildDisposisi(parentID)
	if err != nil {
		return
	}

	// Check if all children are completed or rejected
	allFinished := true
	for _, child := range children {
		if !child.IsCompleted() && child.Status != models.StatusRejected {
			allFinished = false
			break
		}
	}

	// If all finished, mark parent as completed
	if allFinished && len(children) > 0 {
		s.repo.UpdateStatus(parentID, models.StatusCompleted)
	}
}

// convertDisposisiToTreeNode - convert model to DTO tree node
func (s *DisposisiServiceImpl) convertDisposisiToTreeNode(d *models.Disposisi) dto.DisposisiTreeNode {
	node := dto.DisposisiTreeNode{
		ID: d.ID,
		FromUser: dto.UserBasicInfo{
			ID:    d.FromUser.ID,
			Name:  d.FromUser.Name,
			Email: d.FromUser.Email,
		},
		ToUser: dto.UserBasicInfo{
			ID:    d.ToUser.ID,
			Name:  d.ToUser.Name,
			Email: d.ToUser.Email,
		},
		Level:      d.Level,
		Status:     d.Status,
		Catatan:    d.Catatan,
		CreatedAt:  d.CreatedAt,
		CompleteAt: d.CompleteAt,
		Children:   make([]dto.DisposisiTreeNode, 0),
	}

	// Recursive convert children
	if d.ChildDisposisi != nil {
		for _, child := range d.ChildDisposisi {
			childCopy := child // Create copy
			node.Children = append(node.Children, s.convertDisposisiToTreeNode(&childCopy))
		}
	}

	return node
}
