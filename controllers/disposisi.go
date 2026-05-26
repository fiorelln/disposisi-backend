package controllers

import (
	"net/http"
	"time"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/models"
	"github.com/gin-gonic/gin"
)

func CreateDisposisi(c *gin.Context) {

	userID, _ := c.Get("user_id")

	var input struct {
		IDSuratMasuk uint   `json:"id_surat_masuk"`
		IDPenerima   uint   `json:"id_penerima"`
		Sifat        string `json:"sifat"`
		Catatan      string `json:"catatan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	now := time.Now()

	disposisi := models.Disposisi{
		Sifat:        input.Sifat,
		Catatan:      input.Catatan,
		SuratMasukID: input.IDSuratMasuk,
		ToUserID:     input.IDPenerima,
		FromUserID:   userID.(uint),
		Status:       models.StatusPending,
		CreatedAt:    now,
	}

	if err := config.DB.Create(&disposisi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal membuat disposisi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "disposisi berhasil dibuat",
		"data": disposisi,
	})
}

func GetDisposisi(c *gin.Context) {

	var disposisi []models.Disposisi

	if err := config.DB.Find(&disposisi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal mengambil disposisi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": disposisi,
	})
}

func ListDisposisi(c *gin.Context) {

	var disposisi []models.Disposisi

	if err := config.DB.Order("id_disposisi DESC").Find(&disposisi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal mengambil data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": disposisi,
	})
}

func ApproveDisposisi(c *gin.Context) {

	var input struct {
		IDDisposisi uint   `json:"id_disposisi"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var disposisi models.Disposisi

	if err := config.DB.First(&disposisi, "id_disposisi = ?", input.IDDisposisi).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "disposisi tidak ditemukan",
		})
		return
	}

	now := time.Now()

	disposisi.Status = input.Status
	if input.Status == "completed" {
		disposisi.CompleteAt = &now
	}

	config.DB.Save(&disposisi)

	c.JSON(http.StatusOK, gin.H{
		"message": "approval berhasil",
	})
}
