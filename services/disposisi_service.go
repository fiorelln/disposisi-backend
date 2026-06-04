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

type DisposisiService interface {
	CreateInitialDisposisi(suratMasukID uint, toUserID uint, notes string) (*models.Disposisi, error)
	ForwardDisposisi(disposisiID uint, req *dto.CreateForwardRequest, currentUserID uint) (*models.Disposisi, error)
	CompleteDisposisi(disposisiID uint, req *dto.CompleteDisposisiRequest, currentUserID uint) error
	RejectDisposisi(disposisiID uint, reason string, currentUserID uint) error

	MarkAsRead(disposisiID uint) error
	MarkAsReadBatch(disposisiIDs []uint) error

	GetInbox(userID uint, page, pageSize int) (*dto.InboxListResponse, error)
	GetSentItems(userID uint, page, pageSize int) (*dto.SentListResponse, error)
	GetHistory(suratMasukID uint) (*dto.HistoryResponse, error)
	GetDisposisiDetail(disposisiID uint) (*dto.DisposisiResponse, error)

	ValidateForward(disposisiID uint, toUserID uint, currentUserID uint) (bool, string, error)
	CheckCanForward(currentUserID uint, toUserID uint) (bool, error)

	GetStats(userID uint) (map[string]interface{}, error)
}

type DisposisiServiceImpl struct {
	repo            repositories.DisposisiRepository
	permissionSvc   PermissionService
	notificationSvc NotificationService
	db              *gorm.DB
}

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


func (s *DisposisiServiceImpl) CreateInitialDisposisi(
	suratMasukID uint,
	toUserID uint,
	notes string,
) (*models.Disposisi, error) {
	var surat models.SuratMasuk
	if err := s.db.Where("id_surat_masuk = ?", suratMasukID).First(&surat).Error; err != nil {
		return nil, fmt.Errorf("surat not found: %w", err)
	}

	disposisi := &models.Disposisi{
		SuratMasukID:      suratMasukID,
		FromUserID:        1,
		ToUserID:          toUserID,
		ParentDisposisiID: nil,
		Level:             0,
		Status:            models.StatusPending,
		Catatan:           notes,
		Dibaca:            false,
	}

	if err := s.repo.Create(disposisi); err != nil {
		return nil, fmt.Errorf("failed to create disposisi: %w", err)
	}

	s.notificationSvc.NotifyDisposisiReceived(disposisi.ID, toUserID)

	return disposisi, nil
}

func (s *DisposisiServiceImpl) ForwardDisposisi(
	disposisiID uint,
	req *dto.CreateForwardRequest,
	currentUserID uint,
) (*models.Disposisi, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	parentDisposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return nil, fmt.Errorf("parent disposisi not found: %w", err)
	}

	if parentDisposisi.ToUserID != currentUserID {
		return nil, errors.New("unauthorized: you are not the recipient of this disposisi")
	}

	canForward, err := s.CheckCanForward(currentUserID, req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !canForward {
		return nil, errors.New("you don't have permission to forward to this user")
	}

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

	tx := s.db.Begin()

	if err := s.repo.Create(childDisposisi); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create child disposisi: %w", err)
	}

	if err := s.repo.UpdateStatus(disposisiID, models.StatusForwarded); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update parent status: %w", err)
	}

	if !parentDisposisi.Dibaca {
		s.repo.MarkAsRead(disposisiID)
	}

	tx.Commit()

	s.notificationSvc.NotifyDisposisiReceived(childDisposisi.ID, req.ToUserID)

	return childDisposisi, nil
}

func (s *DisposisiServiceImpl) CompleteDisposisi(
	disposisiID uint,
	req *dto.CompleteDisposisiRequest,
	currentUserID uint,
) error {
	disposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return fmt.Errorf("disposisi not found: %w", err)
	}

	if disposisi.ToUserID != currentUserID {
		return errors.New("unauthorized: you are not the recipient of this disposisi")
	}

	if disposisi.IsCompleted() {
		return errors.New("disposisi is already completed")
	}

	updateData := map[string]interface{}{
		"status":                models.StatusCompleted,
		"complete_at":           time.Now(),
		"tanggapan_saran":       req.TanggapanSaran,
		"proses_lanjut":         req.ProsesLanjut,
		"koordinasi_konfirmasi": req.KoordinasiKonfirmasi,
		"catatan":               req.Catatan,
	}

	tx := s.db.Begin()

	if err := s.repo.MarkAsCompleted(disposisiID, updateData); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to complete disposisi: %w", err)
	}

	if disposisi.ParentDisposisiID != nil {
		s.autoUpdateParentStatus(*disposisi.ParentDisposisiID)
	}

	tx.Commit()

	s.notificationSvc.NotifyDisposisiCompleted(disposisiID, disposisi.FromUserID)

	return nil
}

func (s *DisposisiServiceImpl) RejectDisposisi(
	disposisiID uint,
	reason string,
	currentUserID uint,
) error {
	disposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return fmt.Errorf("disposisi not found: %w", err)
	}

	if disposisi.ToUserID != currentUserID {
		return errors.New("unauthorized: you are not the recipient of this disposisi")
	}

	if err := s.repo.Update(&models.Disposisi{
		ID:         disposisiID,
		Status:     models.StatusRejected,
		Catatan:    reason,
		CompleteAt: &time.Time{},
	}); err != nil {
		return fmt.Errorf("failed to reject disposisi: %w", err)
	}

	s.notificationSvc.NotifyDisposisiRejected(disposisiID, disposisi.FromUserID)

	return nil
}


func (s *DisposisiServiceImpl) MarkAsRead(disposisiID uint) error {
	return s.repo.MarkAsRead(disposisiID)
}

func (s *DisposisiServiceImpl) MarkAsReadBatch(disposisiIDs []uint) error {
	for _, id := range disposisiIDs {
		if err := s.repo.MarkAsRead(id); err != nil {
			return err
		}
	}
	return nil
}


func (s *DisposisiServiceImpl) GetInbox(
	userID uint,
	page, pageSize int,
) (*dto.InboxListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	disposisi, total, err := s.repo.GetInbox(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}

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

func (s *DisposisiServiceImpl) GetSentItems(
	userID uint,
	page, pageSize int,
) (*dto.SentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	disposisi, total, err := s.repo.GetSentItems(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent items: %w", err)
	}

	items := make([]dto.SentItemResponse, len(disposisi))
	for i, d := range disposisi {
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

func (s *DisposisiServiceImpl) GetHistory(suratMasukID uint) (*dto.HistoryResponse, error) {
	var surat models.SuratMasuk
	if err := s.db.Where("id_surat_masuk = ?", suratMasukID).First(&surat).Error; err != nil {
		return nil, fmt.Errorf("surat not found: %w", err)
	}

	rootDisposisi, err := s.repo.GetRootDisposisi(suratMasukID)
	if err != nil {
		return nil, fmt.Errorf("root disposisi not found: %w", err)
	}

	treeDisposisi, err := s.repo.GetDisposisiTree(rootDisposisi.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	allDisposisi, err := s.repo.GetHistory(suratMasukID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

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


func (s *DisposisiServiceImpl) ValidateForward(
	disposisiID uint,
	toUserID uint,
	currentUserID uint,
) (bool, string, error) {
	disposisi, err := s.repo.GetByID(disposisiID)
	if err != nil {
		return false, "Disposisi tidak ditemukan", err
	}

	if disposisi.ToUserID != currentUserID {
		return false, "Anda bukan penerima disposisi ini", nil
	}

	if !disposisi.IsPending() {
		return false, fmt.Sprintf("Disposisi sudah dalam status %s", disposisi.Status), nil
	}

	canForward, err := s.CheckCanForward(currentUserID, toUserID)
	if err != nil {
		return false, "Error checking permission", err
	}

	if !canForward {
		return false, "Anda tidak memiliki izin untuk forward ke user tersebut", nil
	}

	return true, "", nil
}

func (s *DisposisiServiceImpl) CheckCanForward(currentUserID uint, toUserID uint) (bool, error) {
	return s.permissionSvc.CanForward(currentUserID, toUserID)
}


func (s *DisposisiServiceImpl) GetStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	unread, err := s.repo.CountUnreadInbox(userID)
	if err != nil {
		return nil, fmt.Errorf("error counting unread: %w", err)
	}
	stats["unread_inbox"] = unread

	pending, err := s.repo.CountPendingInbox(userID)
	if err != nil {
		return nil, fmt.Errorf("error counting pending: %w", err)
	}
	stats["pending_inbox"] = pending

	inbox, _, err := s.repo.GetInbox(userID, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("error getting inbox: %w", err)
	}
	stats["total_inbox"] = len(inbox)

	return stats, nil
}


func (s *DisposisiServiceImpl) autoUpdateParentStatus(parentID uint) {
	children, err := s.repo.GetChildDisposisi(parentID)
	if err != nil {
		return
	}

	allFinished := true
	for _, child := range children {
		if !child.IsCompleted() && child.Status != models.StatusRejected {
			allFinished = false
			break
		}
	}

	if allFinished && len(children) > 0 {
		s.repo.UpdateStatus(parentID, models.StatusCompleted)
	}
}

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

	if d.ChildDisposisi != nil {
		for _, child := range d.ChildDisposisi {
			childCopy := child
			node.Children = append(node.Children, s.convertDisposisiToTreeNode(&childCopy))
		}
	}

	return node
}
