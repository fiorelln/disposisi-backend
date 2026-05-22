package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/models"
	"github.com/gin-gonic/gin"
)

func UploadSurat(c *gin.Context) {

	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file wajib diupload",
		})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file harus pdf",
		})
		return
	}

	var input struct {
		NoSurat      string `form:"no_surat"`
		PerihalSurat string `form:"perihal_surat"`
		AsalSurat    string `form:"asal_surat"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	uploadPath := "uploads/surat"

	os.MkdirAll(uploadPath, os.ModePerm)

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)

	path := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal upload file",
		})
		return
	}

	now := time.Now()

	surat := models.SuratMasuk{
		NoSurat:         input.NoSurat,
		PerihalSurat:    input.PerihalSurat,
		AsalSurat:       input.AsalSurat,
		FilePDF:         path,
		TanggalDiterima: &now,
		StatusAlur:      "masuk",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := config.DB.Create(&surat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal simpan database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "upload surat berhasil",
		"data": surat,
	})
}

func GetSurat(c *gin.Context) {

	id := c.Param("surat_id")

	var surat models.SuratMasuk

	if err := config.DB.First(&surat, "id_surat_masuk = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "surat tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": surat,
	})
}

func ListSurat(c *gin.Context) {

	var surat []models.SuratMasuk

	if err := config.DB.Order("id_surat_masuk DESC").Find(&surat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal mengambil surat",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": surat,
	})
}
