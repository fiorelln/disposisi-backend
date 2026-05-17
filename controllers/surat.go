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

	userID, _ := c.Get("user_id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File wajib diupload"})
		return
	}

	// Validasi tipe file
	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File harus PDF"})
		return
	}

	// Validasi ukuran file (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file maksimal 5MB"})
		return
	}

	// Validasi MIME type
	if file.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MIME type harus PDF"})
		return
	}

	// Ambil input form
	var input struct {
		Kategori  string `form:"kategori" binding:"required"`
		Judul     string `form:"judul" binding:"required"`
		Deskripsi string `form:"deskripsi"`
		TujuanID  uint   `form:"tujuan_id" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploadPath := "uploads/surat"
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat folder upload"})
		return
	}

	// Sanitasi nama file
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	path := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload file"})
		return
	}

	surat := models.Surat{
		FileSurat: path,
		Status:    "dikirim",
		PengirimID: uint(userID.(float64)),
		TujuanID:   input.TujuanID,
		Kategori:   input.Kategori,
		Judul:      input.Judul,
		Deskripsi:  input.Deskripsi,
	}

	if err := config.DB.Create(&surat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB insert gagal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload surat berhasil",
		"surat": surat,
	})
}

// GetSurat untuk melihat detail surat
func GetSurat(c *gin.Context) {

	suratID := c.Param("surat_id")

	var surat models.Surat

	if err := config.DB.
		Preload("Pengirim").
		Preload("Tujuan").
		Preload("Disposisi").
		First(&surat, suratID).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "Surat tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": surat,
	})
}

// ListSurat untuk melihat daftar surat
func ListSurat(c *gin.Context) {

	userID, _ := c.Get("user_id")
	kategori := c.Query("kategori")
	status := c.Query("status")

	var surats []models.Surat

	query := config.DB.
		Preload("Pengirim").
		Preload("Tujuan").
		Preload("Disposisi").
		Where("tujuan_id = ? OR pengirim_id = ?", uint(userID.(float64)), uint(userID.(float64)))

	if kategori != "" {
		query = query.Where("kategori = ?", kategori)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&surats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": surats,
	})
}