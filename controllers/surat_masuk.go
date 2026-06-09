package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/helpers"
	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/services"
	"github.com/gin-gonic/gin"
)

const maxFileSize = 10 << 20

type SuratMasukController struct {
	service      services.SuratMasukService
	disposisiSvc services.DisposisiService
}

func NewSuratMasukController(service services.SuratMasukService, disposisiSvc services.DisposisiService) *SuratMasukController {
	return &SuratMasukController{service: service, disposisiSvc: disposisiSvc}
}

func (ctrl *SuratMasukController) Register(c *gin.Context) {
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
	perihal := c.PostForm("perihal_surat")
	asal := c.PostForm("asal_surat")

	if noSurat == "" || perihal == "" || asal == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nomor surat, perihal, dan asal surat wajib diisi"})
		return
	}

	uploadPath := "uploads/surat_masuk"
	os.MkdirAll(uploadPath, os.ModePerm)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	path := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan file ke server"})
		return
	}

	surat, err := ctrl.service.Register(noSurat, perihal, asal, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	helpers.CompressFile(path)

	c.JSON(http.StatusOK, gin.H{"message": "surat masuk berhasil didaftarkan", "data": surat})
}

func (ctrl *SuratMasukController) ForwardToPrincipal(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	userID := c.MustGet("user_id").(uint)

	if err := ctrl.service.ForwardToPrincipal(uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "surat berhasil diteruskan ke Kepala Sekolah"})
}

func (ctrl *SuratMasukController) Review(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	userID := c.MustGet("user_id").(uint)

	var input struct {
		StatusApproval string `json:"status_approval"`
		Catatan        string `json:"catatan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format data tidak valid"})
		return
	}

	if input.StatusApproval != "disetujui" && input.StatusApproval != "ditolak" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status_approval harus 'disetujui' atau 'ditolak'"})
		return
	}

	if err := ctrl.service.Review(uint(id), userID, input.StatusApproval, input.Catatan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := "surat berhasil disetujui"
	if input.StatusApproval == "ditolak" {
		msg = "surat ditolak"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func (ctrl *SuratMasukController) DistributeToUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	userID := c.MustGet("user_id").(uint)

	var input struct {
		PenerimaID        uint   `json:"penerima_id"`
		JabatanPenerimaID uint   `json:"jabatan_penerima_id"`
		Catatan           string `json:"catatan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data penerima tidak valid"})
		return
	}

	if input.PenerimaID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "penerima_id wajib diisi"})
		return
	}

	disposisi, err := ctrl.service.DistributeToUser(uint(id), userID, input.PenerimaID, input.JabatanPenerimaID, input.Catatan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "surat berhasil didistribusikan", "data": disposisi})
}

func (ctrl *SuratMasukController) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	surat, err := ctrl.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "data surat tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": surat})
}

func (ctrl *SuratMasukController) Download(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	surat, err := ctrl.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "surat tidak ditemukan"})
		return
	}

	c.FileAttachment(surat.FilePDF, filepath.Base(surat.FilePDF))
}

func (ctrl *SuratMasukController) List(c *gin.Context) {
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

func (ctrl *SuratMasukController) GetInboxDistribusi(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, total, err := ctrl.disposisiSvc.GetInbox(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memuat inbox"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

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
