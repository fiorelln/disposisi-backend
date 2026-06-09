package controllers

import (
	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/models"
	"github.com/gin-gonic/gin"
)

func Dashboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	roles, _ := c.Get("roles")

	var totalSuratMasuk int64
	config.DB.Model(&models.SuratMasuk{}).Count(&totalSuratMasuk)

	var totalSuratKeluar int64
	config.DB.Model(&models.SuratKeluar{}).Count(&totalSuratKeluar)

	var totalDisposisi int64
	config.DB.Model(&models.Disposisi{}).
		Where("id_penerima = ?", uid).Count(&totalDisposisi)

	var pendingDisposisi int64
	config.DB.Model(&models.Disposisi{}).
		Where("id_penerima = ? AND status_disposisi = ?", uid, "belum_dibaca").Count(&pendingDisposisi)

	var totalSuratMasukMenunggu int64
	config.DB.Model(&models.SuratMasuk{}).
		Where("status_alur = ?", "disposisi_kepsek").Count(&totalSuratMasukMenunggu)

	c.JSON(200, gin.H{
		"roles": roles,
		"summary": gin.H{
			"total_surat_masuk":          totalSuratMasuk,
			"total_surat_keluar":         totalSuratKeluar,
			"total_disposisi":            totalDisposisi,
			"pending_disposisi":          pendingDisposisi,
			"surat_masuk_menunggu":       totalSuratMasukMenunggu,
		},
	})
}
