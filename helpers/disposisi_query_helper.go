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

func (h *DisposisiQueryHelper) GetDisposisiStatsForUser(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalInbox int64
	h.db.Where("id_penerima = ?", userID).
		Model(&models.Disposisi{}).
		Count(&totalInbox)
	stats["total_inbox"] = totalInbox

	var unread int64
	h.db.Where("id_penerima = ? AND status_disposisi = ?", userID, "belum_dibaca").
		Model(&models.Disposisi{}).
		Count(&unread)
	stats["belum_dibaca"] = unread

	var selesai int64
	h.db.Where("id_penerima = ? AND status_disposisi = ?", userID, "selesai").
		Model(&models.Disposisi{}).
		Count(&selesai)
	stats["selesai"] = selesai

	return stats, nil
}
