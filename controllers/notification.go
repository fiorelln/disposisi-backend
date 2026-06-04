package controllers

import (
	"net/http"
	"strconv"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/models"
	"github.com/gin-gonic/gin"
)

func (ctrl *SuratMasukController) ListNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	var notifs []models.Notifikasi
	config.DB.Model(&models.Notifikasi{}).Where("id_penerima = ?", userID).Count(&total)
	config.DB.Where("id_penerima = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&notifs)

	items := make([]gin.H, 0, len(notifs))
	for _, n := range notifs {
		items = append(items, gin.H{
			"id":         n.IDNotifikasi,
			"jenis":      n.Jenis,
			"judul":      n.Judul,
			"pesan":      n.Pesan,
			"is_read":    n.IsRead,
			"created_at": n.CreatedAt,
		})
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

func (ctrl *SuratMasukController) MarkNotificationRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	result := config.DB.Model(&models.Notifikasi{}).
		Where("id_notifikasi = ? AND id_penerima = ?", id, userID).
		Update("is_read", true)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "notifikasi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notifikasi telah ditandai dibaca"})
}

func (ctrl *SuratMasukController) CountUnreadNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var count int64
	config.DB.Model(&models.Notifikasi{}).
		Where("id_penerima = ? AND is_read = ?", userID, false).
		Count(&count)
	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
