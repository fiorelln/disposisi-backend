package controllers

import (
	"net/http"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/models"
	"github.com/gin-gonic/gin"
)

func CreateDisposisi(c *gin.Context) {

	userID, _ := c.Get("user_id")

	var input struct {
		SuratID   uint   `json:"surat_id" binding:"required"`
		TujuanID  uint   `json:"tujuan_id" binding:"required"`
		Tujuan    string `json:"tujuan" binding:"required"`
		Catatan   string `json:"catatan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var surat models.Surat
	if err := config.DB.First(&surat, input.SuratID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Surat tidak ditemukan"})
		return
	}

	disposisi := models.Disposisi{
		SuratID:          input.SuratID,
		TujuanID:         input.TujuanID,
		Tujuan:           input.Tujuan,
		Catatan:          input.Catatan,
		Status:           "menunggu",
		VerifikasiStatus: "menunggu",
	}

	if err := config.DB.Create(&disposisi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat disposisi"})
		return
	}

	config.DB.Model(&surat).Update("status", "diteruskan")

	c.JSON(http.StatusOK, gin.H{
		"message": "Disposisi berhasil dibuat",
		"disposisi": disposisi,
	})
}

func ApproveDisposisi(c *gin.Context) {

	userID, _ := c.Get("user_id")

	var input struct {
		DisposisiID uint   `json:"disposisi_id" binding:"required"`
		IsApproved  bool   `json:"is_approved" binding:"required"`
		Catatan     string `json:"catatan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var disposisi models.Disposisi

	if err := config.DB.First(&disposisi, input.DisposisiID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Disposisi tidak ditemukan"})
		return
	}

	var surat models.Surat
	config.DB.First(&surat, disposisi.SuratID)

	if input.IsApproved {
		disposisi.VerifikasiStatus = "setuju"
		disposisi.Status = "disetujui"
		surat.Status = "disetujui"
	} else {
		disposisi.VerifikasiStatus = "tolak"
		disposisi.Status = "ditolak"
		surat.Status = "ditolak"
	}

	disposisi.VerifikatorID = uint(userID.(float64))
	if input.Catatan != "" {
		disposisi.Catatan = input.Catatan
	}

	config.DB.Save(&disposisi)
	config.DB.Save(&surat)

	c.JSON(http.StatusOK, gin.H{
		"message": "Disposisi berhasil diproses",
		"disposisi": disposisi,
	})
}

func GetDisposisi(c *gin.Context) {

	suratID := c.Param("surat_id")

	var disposisi models.Disposisi

	if err := config.DB.
		Preload("Surat").
		Preload("TujuanUser").
		Preload("Verifikator").
		Where("surat_id = ?", suratID).
		First(&disposisi).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "Disposisi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": disposisi,
	})
}

func ListDisposisi(c *gin.Context) {

	userID, _ := c.Get("user_id")
	status := c.Query("status") // optional filter

	var disposisi []models.Disposisi
	query := config.DB.
		Preload("Surat").
		Preload("TujuanUser").
		Preload("Verifikator")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query = query.Where("tujuan_id = ? OR verifikator_id = ?", uint(userID.(float64)), uint(userID.(float64)))

	if err := query.Find(&disposisi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": disposisi,
	})
}
