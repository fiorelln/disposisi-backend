package helpers

import (
	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type DisposisiQueryHelper struct {
	db *gorm.DB
}

func NewDisposisiQueryHelper(db *gorm.DB) *DisposisiQueryHelper {
	return &DisposisiQueryHelper{db: db}
}


type InboxQueryBuilder struct {
	db         *gorm.DB
	userID     uint
	status     string
	page       int
	pageSize   int
	sortBy     string
	sortOrder  string
	includeRead bool
}

func (h *DisposisiQueryHelper) NewInboxQueryBuilder(userID uint) *InboxQueryBuilder {
	return &InboxQueryBuilder{
		db:       h.db,
		userID:   userID,
		page:     1,
		pageSize: 20,
		sortBy:   "created_at",
		sortOrder: "DESC",
	}
}

func (q *InboxQueryBuilder) WithStatus(status string) *InboxQueryBuilder {
	q.status = status
	return q
}

func (q *InboxQueryBuilder) Unread() *InboxQueryBuilder {
	q.includeRead = false
	q.db = q.db.Where("dibaca = ?", false)
	return q
}

func (q *InboxQueryBuilder) WithPage(page, pageSize int) *InboxQueryBuilder {
	if page > 0 {
		q.page = page
	}
	if pageSize > 0 && pageSize <= 100 {
		q.pageSize = pageSize
	}
	return q
}

func (q *InboxQueryBuilder) SortBy(field string) *InboxQueryBuilder {
	allowedFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"dibaca":     true,
		"status":     true,
	}
	if allowedFields[field] {
		q.sortBy = field
	}
	return q
}

func (q *InboxQueryBuilder) SortAsc() *InboxQueryBuilder {
	q.sortOrder = "ASC"
	return q
}

func (q *InboxQueryBuilder) Build() ([]models.Disposisi, int64, error) {
	var disposisi []models.Disposisi
	var total int64

	query := q.db.Where("to_user_id = ? AND deleted_at IS NULL", q.userID)

	if q.status != "" {
		query = query.Where("status = ?", q.status)
	}

	if err := query.Model(&models.Disposisi{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (q.page - 1) * q.pageSize
	err := query.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		Order(q.sortBy + " " + q.sortOrder).
		Offset(offset).
		Limit(q.pageSize).
		Find(&disposisi).Error

	return disposisi, total, err
}


func (h *DisposisiQueryHelper) GetDisposisiByStatus(status string, userID uint, isInbox bool) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi

	query := h.db.Where("status = ? AND deleted_at IS NULL", status)

	if isInbox {
		query = query.Where("to_user_id = ?", userID)
	} else {
		query = query.Where("from_user_id = ?", userID)
	}

	err := query.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		Order("created_at DESC").
		Find(&disposisi).Error

	return disposisi, err
}

func (h *DisposisiQueryHelper) GetRecentDisposisi(userID uint, limit int, isInbox bool) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi

	query := h.db.Where("deleted_at IS NULL")

	if isInbox {
		query = query.Where("to_user_id = ?", userID)
	} else {
		query = query.Where("from_user_id = ?", userID)
	}

	err := query.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		Order("created_at DESC").
		Limit(limit).
		Find(&disposisi).Error

	return disposisi, err
}

func (h *DisposisiQueryHelper) GetDisposisiByPriority(sifat string, userID uint, isInbox bool) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi

	query := h.db.Where("sifat = ? AND deleted_at IS NULL", sifat)

	if isInbox {
		query = query.Where("to_user_id = ?", userID)
	} else {
		query = query.Where("from_user_id = ?", userID)
	}

	err := query.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		Order("created_at DESC").
		Find(&disposisi).Error

	return disposisi, err
}


func (h *DisposisiQueryHelper) GetDisposisiStatsForUser(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalInbox int64
	h.db.Where("to_user_id = ? AND deleted_at IS NULL", userID).
		Model(&models.Disposisi{}).
		Count(&totalInbox)
	stats["total_inbox"] = totalInbox

	var totalSent int64
	h.db.Where("from_user_id = ? AND deleted_at IS NULL", userID).
		Model(&models.Disposisi{}).
		Count(&totalSent)
	stats["total_sent"] = totalSent

	var unread int64
	h.db.Where("to_user_id = ? AND dibaca = ? AND deleted_at IS NULL", userID, false).
		Model(&models.Disposisi{}).
		Count(&unread)
	stats["unread"] = unread

	var pending int64
	h.db.Where("to_user_id = ? AND status = ? AND deleted_at IS NULL", userID, models.StatusPending).
		Model(&models.Disposisi{}).
		Count(&pending)
	stats["pending"] = pending

	var statusDist []map[string]interface{}
	h.db.
		Where("to_user_id = ? AND deleted_at IS NULL", userID).
		Model(&models.Disposisi{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusDist)
	stats["status_distribution"] = statusDist

	var priorityDist []map[string]interface{}
	h.db.
		Where("to_user_id = ? AND deleted_at IS NULL", userID).
		Model(&models.Disposisi{}).
		Select("sifat, COUNT(*) as count").
		Group("sifat").
		Scan(&priorityDist)
	stats["priority_distribution"] = priorityDist

	return stats, nil
}


func (h *DisposisiQueryHelper) GetFullChain(disposisiID uint) ([]models.Disposisi, error) {
	var chain []models.Disposisi

	var current models.Disposisi
	if err := h.db.First(&current, disposisiID).Error; err != nil {
		return nil, err
	}

	chain = append(chain, current)

	for current.ParentDisposisiID != nil {
		var parent models.Disposisi
		if err := h.db.First(&parent, *current.ParentDisposisiID).Error; err != nil {
			break
		}
		chain = append(chain, parent)
		current = parent
	}

	return chain, nil
}

func (h *DisposisiQueryHelper) GetBrothers(disposisiID uint) ([]models.Disposisi, error) {
	var disposisi models.Disposisi
	var brothers []models.Disposisi

	if err := h.db.First(&disposisi, disposisiID).Error; err != nil {
		return nil, err
	}

	if disposisi.ParentDisposisiID == nil {
		return brothers, nil
	}

	err := h.db.
		Where("parent_disposisi_id = ? AND deleted_at IS NULL", disposisi.ParentDisposisiID).
		Preload("FromUser").
		Preload("ToUser").
		Order("created_at ASC").
		Find(&brothers).Error

	return brothers, err
}

func (h *DisposisiQueryHelper) GetDescendants(rootID uint) ([]models.Disposisi, error) {
	var descendants []models.Disposisi

	err := h.db.Raw(`
		WITH RECURSIVE disposisi_tree AS (
			SELECT id, parent_disposisi_id, level, from_user_id, to_user_id, status, created_at
			FROM disposisi
			WHERE id = ? AND deleted_at IS NULL
			
			UNION ALL
			
			SELECT d.id, d.parent_disposisi_id, d.level, d.from_user_id, d.to_user_id, d.status, d.created_at
			FROM disposisi d
			INNER JOIN disposisi_tree dt ON d.parent_disposisi_id = dt.id
			WHERE d.deleted_at IS NULL
		)
		SELECT * FROM disposisi_tree
		ORDER BY level, created_at
	`, rootID).Scan(&descendants).Error

	return descendants, err
}


func (h *DisposisiQueryHelper) ArchiveOldDisposisi(days int) (int64, error) {
	result := h.db.
		Where("created_at < NOW() - INTERVAL ? DAY", days).
		Update("deleted_at", gorm.Expr("NOW()"))

	return result.RowsAffected, result.Error
}

func (h *DisposisiQueryHelper) GetDeletedDisposisi() ([]models.Disposisi, error) {
	var disposisi []models.Disposisi

	err := h.db.
		Where("deleted_at IS NOT NULL").
		Preload("FromUser").
		Preload("ToUser").
		Order("deleted_at DESC").
		Find(&disposisi).Error

	return disposisi, err
}


func (h *DisposisiQueryHelper) SearchDisposisi(keyword string, userID uint, isInbox bool) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi

	query := h.db.
		Joins("LEFT JOIN surat_masuk ON disposisi.surat_masuk_id = surat_masuk.id_surat_masuk").
		Joins("LEFT JOIN users u1 ON disposisi.from_user_id = u1.id_user").
		Joins("LEFT JOIN users u2 ON disposisi.to_user_id = u2.id_user").
		Where("disposisi.deleted_at IS NULL")

	if isInbox {
		query = query.Where("disposisi.to_user_id = ?", userID)
	} else {
		query = query.Where("disposisi.from_user_id = ?", userID)
	}

	query = query.Where(
		"disposisi.catatan ILIKE ? OR surat_masuk.no_surat ILIKE ? OR surat_masuk.perihal_surat ILIKE ? OR u1.nama ILIKE ? OR u2.nama ILIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%",
	)

	err := query.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		Order("disposisi.created_at DESC").
		Find(&disposisi).Error

	return disposisi, err
}
