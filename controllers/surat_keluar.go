package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fiorelln/disposisi/services"
	"github.com/gin-gonic/gin"
)

type SuratKeluarController struct {
	service services.SuratKeluarService
}

func NewSuratKeluarController(service services.SuratKeluarService) *SuratKeluarController {
	return &SuratKeluarController{service: service}
}

func (ctrl *SuratKeluarController) Create(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file surat wajib diunggah"})
		return
	}

	if filepath.Ext(file.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format file harus berupa PDF"})
		return
	}

	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran file maksimal 1 MB"})
		return
	}

	noSurat := c.PostForm("no_surat")
	perihal := c.PostForm("perihal")
	catatan := c.PostForm("catatan")
	tujuan := c.PostForm("tujuan")

	if noSurat == "" || perihal == "" || tujuan == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nomor surat, perihal, dan tujuan wajib diisi"})
		return
	}

	uploadPath := "uploads/surat_keluar"
	os.MkdirAll(uploadPath, os.ModePerm)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	path := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan file ke server"})
		return
	}

	surat, err := ctrl.service.Create(noSurat, perihal, catatan, tujuan, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "draft surat keluar berhasil dibuat", "data": surat})
}

func (ctrl *SuratKeluarController) SubmitToPrincipal(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ctrl.service.SubmitToPrincipal(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "surat berhasil diajukan ke Kepala Sekolah"})
}

func (ctrl *SuratKeluarController) Review(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	userID := c.MustGet("user_id").(uint)

	var input struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format data tidak valid"})
		return
	}

	if err := ctrl.service.Review(uint(id), userID, input.Status, input.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "surat keluar berhasil ditinjau"})
}

func (ctrl *SuratKeluarController) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	surat, err := ctrl.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "data surat tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": surat})
}

func (ctrl *SuratKeluarController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	data, total, err := ctrl.service.List(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil daftar surat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"total": total,
		"page":  page,
	})
}
