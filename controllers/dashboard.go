package controllers

import (
	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/models"
	"github.com/gin-gonic/gin"
)

func Dashboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rolesInterface, _ := c.Get("roles")

	userRoles := []string{}
	if roles, ok := rolesInterface.([]interface{}); ok {
		for _, role := range roles {
			if r, ok := role.(string); ok {
				userRoles = append(userRoles, r)
			}
		}
	}

	uid := uint(userID.(float64))

	var totalSurat int64
	config.DB.Model(&models.Surat{}).
		Where("tujuan_id = ? OR pengirim_id = ?", uid, uid).
		Count(&totalSurat)

	var totalDisposisi int64
	config.DB.Model(&models.Disposisi{}).
		Where("tujuan_id = ? OR verifikator_id = ?", uid, uid).
		Count(&totalDisposisi)

	var pendingDisposisi int64
	config.DB.Model(&models.Disposisi{}).
		Where("status = ? AND (tujuan_id = ? OR verifikator_id = ?)", "menunggu", uid, uid).
		Count(&pendingDisposisi)

	var approvalPending int64
	if hasRole(userRoles, "Kepala Sekolah") {
		config.DB.Model(&models.Disposisi{}).
			Where("status = ?", "menunggu").
			Count(&approvalPending)
	}

	c.JSON(200, gin.H{
		"user_id":           uid,
		"roles":             userRoles,
		"summary": gin.H{
			"total_surat":       totalSurat,
			"total_disposisi":   totalDisposisi,
			"pending_disposisi": pendingDisposisi,
			"approval_pending":  approvalPending,
		},
	})
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
