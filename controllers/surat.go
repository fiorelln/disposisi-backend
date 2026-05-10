package controllers

import (
	"fmt"
	"net/http"
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
			"error": "File wajib diupload",
		})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File harus PDF",
		})
		return
	}

	filename := fmt.Sprintf(
		"%d_%s",
		time.Now().Unix(),
		file.Filename,
	)

	path := "uploads/surat/" + filename

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal upload file",
		})
		return
	}

	surat := models.Surat{
		FileSurat: path,
		Status:    "dikirim",
	}

	config.DB.Create(&surat)

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload surat berhasil",
		"file":    path,
	})
}
