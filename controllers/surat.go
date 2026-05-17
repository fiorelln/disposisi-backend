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
		c.JSON(http.StatusBadRequest, gin.H{"error": "File wajib diupload"})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File harus PDF"})
		return
	}

	uploadPath := "uploads/surat"
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat folder upload"})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	path := uploadPath + "/" + filename

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload file"})
		return
	}

	surat := models.Surat{
		FileSurat: path,
		Status:    "dikirim",
	}

	if err := config.DB.Create(&surat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB insert gagal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload surat berhasil",
		"file":    path,
	})
}