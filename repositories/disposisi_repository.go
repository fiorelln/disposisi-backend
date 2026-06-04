package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type DisposisiRepository interface {
	Create(disposisi *models.Disposisi) error
	GetByID(id uint) (*models.Disposisi, error)
	Update(disposisi *models.Disposisi) error
	Delete(id uint) error
	Restore(id uint) error

	GetInbox(userID uint, page, pageSize int) ([]models.Disposisi, int64, error)
	GetSentItems(userID uint, page, pageSize int) ([]models.Disposisi, int64, error)
	GetHistory(suratMasukID uint) ([]models.Disposisi, error)
	GetChildDisposisi(parentID uint) ([]models.Disposisi, error)
	GetRootDisposisi(suratMasukID uint) (*models.Disposisi, error)

	CountUnreadInbox(userID uint) (int64, error)
	CountPendingInbox(userID uint) (int64, error)

	UpdateStatus(id uint, status string) error
	MarkAsRead(id uint) error
	MarkAsCompleted(id uint, data map[string]interface{}) error

	GetDisposisiTree(rootID uint) (*models.Disposisi, error)
	CountCompletedChildren(parentID uint) (int, error)

	GetDisposisiByIDs(ids []uint) ([]models.Disposisi, error)
	UpdateStatusBatch(ids []uint, status string) error
}

type DisposisiRepositoryImpl struct {
	db *gorm.DB
}

func NewDisposisiRepository(db *gorm.DB) DisposisiRepository {
	return &DisposisiRepositoryImpl{db: db}
}


func (r *DisposisiRepositoryImpl) Create(disposisi *models.Disposisi) error {
	if disposisi == nil {
		return errors.New("disposisi cannot be nil")
	}
	return r.db.Create(disposisi).Error
}

func (r *DisposisiRepositoryImpl) GetByID(id uint) (*models.Disposisi, error) {
	var disposisi models.Disposisi
	err := r.db.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&disposisi).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("disposisi not found")
		}
		return nil, err
	}
	return &disposisi, nil
}

func (r *DisposisiRepositoryImpl) Update(disposisi *models.Disposisi) error {
	if disposisi == nil {
		return errors.New("disposisi cannot be nil")
	}
	return r.db.
		Where("id = ? AND deleted_at IS NULL", disposisi.ID).
		Save(disposisi).Error
}

func (r *DisposisiRepositoryImpl) Delete(id uint) error {
	return r.db.
		Model(&models.Disposisi{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

func (r *DisposisiRepositoryImpl) Restore(id uint) error {
	return r.db.
		Model(&models.Disposisi{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}


func (r *DisposisiRepositoryImpl) GetInbox(userID uint, page, pageSize int) ([]models.Disposisi, int64, error) {
	var disposisi []models.Disposisi
	var total int64

	offset := (page - 1) * pageSize

	countErr := r.db.
		Where("to_user_id = ? AND deleted_at IS NULL", userID).
		Model(&models.Disposisi{}).
		Count(&total).Error

	if countErr != nil {
		return nil, 0, countErr
	}

	err := r.db.
		Where("to_user_id = ? AND deleted_at IS NULL", userID).
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&disposisi).Error

	if err != nil {
		return nil, 0, err
	}

	return disposisi, total, nil
}

func (r *DisposisiRepositoryImpl) GetSentItems(userID uint, page, pageSize int) ([]models.Disposisi, int64, error) {
	var disposisi []models.Disposisi
	var total int64

	offset := (page - 1) * pageSize

	countErr := r.db.
		Where("from_user_id = ? AND deleted_at IS NULL", userID).
		Model(&models.Disposisi{}).
		Count(&total).Error

	if countErr != nil {
		return nil, 0, countErr
	}

	err := r.db.
		Where("from_user_id = ? AND deleted_at IS NULL", userID).
		Preload("ToUser").
		Preload("SuratMasuk").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&disposisi).Error

	if err != nil {
		return nil, 0, err
	}

	return disposisi, total, nil
}

func (r *DisposisiRepositoryImpl) GetHistory(suratMasukID uint) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi
	err := r.db.
		Where("surat_masuk_id = ? AND deleted_at IS NULL", suratMasukID).
		Preload("FromUser").
		Preload("ToUser").
		Order("level ASC, created_at ASC").
		Find(&disposisi).Error

	return disposisi, err
}

func (r *DisposisiRepositoryImpl) GetChildDisposisi(parentID uint) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi
	err := r.db.
		Where("parent_disposisi_id = ? AND deleted_at IS NULL", parentID).
		Preload("FromUser").
		Preload("ToUser").
		Order("created_at ASC").
		Find(&disposisi).Error

	return disposisi, err
}

func (r *DisposisiRepositoryImpl) GetRootDisposisi(suratMasukID uint) (*models.Disposisi, error) {
	var disposisi models.Disposisi
	err := r.db.
		Where("surat_masuk_id = ? AND level = 0 AND deleted_at IS NULL", suratMasukID).
		Preload("FromUser").
		Preload("ToUser").
		First(&disposisi).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("root disposisi not found")
		}
		return nil, err
	}
	return &disposisi, nil
}


func (r *DisposisiRepositoryImpl) CountUnreadInbox(userID uint) (int64, error) {
	var count int64
	err := r.db.
		Where("to_user_id = ? AND dibaca = ? AND deleted_at IS NULL", userID, false).
		Model(&models.Disposisi{}).
		Count(&count).Error

	return count, err
}

func (r *DisposisiRepositoryImpl) CountPendingInbox(userID uint) (int64, error) {
	var count int64
	err := r.db.
		Where("to_user_id = ? AND status = ? AND deleted_at IS NULL", userID, models.StatusPending).
		Model(&models.Disposisi{}).
		Count(&count).Error

	return count, err
}


func (r *DisposisiRepositoryImpl) UpdateStatus(id uint, status string) error {
	return r.db.
		Model(&models.Disposisi{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("status", status).
		Update("updated_at", time.Now()).Error
}

func (r *DisposisiRepositoryImpl) MarkAsRead(id uint) error {
	return r.db.
		Model(&models.Disposisi{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"dibaca":     true,
			"baca_at":    time.Now(),
			"updated_at": time.Now(),
		}).Error
}

func (r *DisposisiRepositoryImpl) MarkAsCompleted(id uint, data map[string]interface{}) error {
	data["status"] = models.StatusCompleted
	data["complete_at"] = time.Now()
	data["updated_at"] = time.Now()

	return r.db.
		Model(&models.Disposisi{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(data).Error
}


func (r *DisposisiRepositoryImpl) GetDisposisiTree(rootID uint) (*models.Disposisi, error) {
	var disposisi models.Disposisi

	err := r.db.
		Where("id = ? AND deleted_at IS NULL", rootID).
		Preload("FromUser").
		Preload("ToUser").
		Preload("ChildDisposisi", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL")
		}).
		First(&disposisi).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("disposisi not found")
		}
		return nil, err
	}

	r.loadDisposisiTreeRecursive(&disposisi, 0, 5)

	return &disposisi, nil
}

func (r *DisposisiRepositoryImpl) loadDisposisiTreeRecursive(disposisi *models.Disposisi, level, maxLevel int) {
	if level >= maxLevel || disposisi == nil {
		return
	}

	for i := range disposisi.ChildDisposisi {
		child := &disposisi.ChildDisposisi[i]

		r.db.
			Preload("FromUser").
			Preload("ToUser").
			Preload("ChildDisposisi", func(db *gorm.DB) *gorm.DB {
				return db.Where("deleted_at IS NULL")
			}).
			First(child)

		r.loadDisposisiTreeRecursive(child, level+1, maxLevel)
	}
}

func (r *DisposisiRepositoryImpl) CountCompletedChildren(parentID uint) (int, error) {
	var count int64
	err := r.db.
		Where("parent_disposisi_id = ? AND status = ? AND deleted_at IS NULL", parentID, models.StatusCompleted).
		Model(&models.Disposisi{}).
		Count(&count).Error

	return int(count), err
}


func (r *DisposisiRepositoryImpl) GetDisposisiByIDs(ids []uint) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi
	err := r.db.
		Where("id IN ? AND deleted_at IS NULL", ids).
		Preload("FromUser").
		Preload("ToUser").
		Find(&disposisi).Error

	return disposisi, err
}

func (r *DisposisiRepositoryImpl) UpdateStatusBatch(ids []uint, status string) error {
	return r.db.
		Model(&models.Disposisi{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}


func (r *DisposisiRepositoryImpl) GetDisposisiChain(childID uint) ([]models.Disposisi, error) {
	var chain []models.Disposisi

	disposisi, err := r.GetByID(childID)
	if err != nil {
		return nil, err
	}

	chain = append(chain, *disposisi)

	for disposisi.ParentDisposisiID != nil {
		parent, err := r.GetByID(*disposisi.ParentDisposisiID)
		if err != nil {
			break
		}
		chain = append(chain, *parent)
		disposisi = parent
	}

	return chain, nil
}

func (r *DisposisiRepositoryImpl) CheckUserExists(userID uint) (bool, error) {
	var count int64
	err := r.db.
		Where("(from_user_id = ? OR to_user_id = ?) AND deleted_at IS NULL", userID, userID).
		Model(&models.Disposisi{}).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *DisposisiRepositoryImpl) GetDisposisiStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	unread, err := r.CountUnreadInbox(userID)
	if err != nil {
		return nil, fmt.Errorf("error counting unread: %w", err)
	}
	stats["unread_inbox"] = unread

	pending, err := r.CountPendingInbox(userID)
	if err != nil {
		return nil, fmt.Errorf("error counting pending: %w", err)
	}
	stats["pending_inbox"] = pending

	inbox, _, err := r.GetInbox(userID, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("error getting inbox: %w", err)
	}
	stats["total_inbox"] = len(inbox)

	return stats, nil
}
