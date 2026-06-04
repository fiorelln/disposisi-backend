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

const maxFileSize = 1 << 20

type SuratMasukController struct {
	service      services.SuratMasukService
	distribusiSvc services.DistribusiSMService
}

func NewSuratMasukController(service services.SuratMasukService, distribusiSvc services.DistribusiSMService) *SuratMasukController {
	return &SuratMasukController{service: service, distribusiSvc: distribusiSvc}
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

	c.JSON(http.StatusOK, gin.H{"message": "surat masuk berhasil didaftarkan", "data": surat})
}

func (ctrl *SuratMasukController) ForwardToPrincipal(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ctrl.service.ForwardToPrincipal(uint(id)); err != nil {
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
		StatusApproval      string `json:"status_approval"`
		Catatan             string `json:"catatan"`
		TanggapanSaran      string `json:"tanggapan_saran"`
		ProsesLanjut        string `json:"proses_lanjut"`
		KoordinasiKonfirmasi string `json:"koordinasi_konfirmasi"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format data tidak valid"})
		return
	}

	if err := ctrl.service.Review(uint(id), userID, input.StatusApproval, input.Catatan, input.TanggapanSaran, input.ProsesLanjut, input.KoordinasiKonfirmasi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "hasil tinjauan disposisi berhasil disimpan"})
}

func (ctrl *SuratMasukController) Distribute(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var input struct {
		JabatanIDs []uint `json:"jabatan_ids"`
		Catatan    string `json:"catatan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "daftar jabatan tujuan tidak valid"})
		return
	}

	if err := ctrl.distribusiSvc.Distribute(uint(id), input.JabatanIDs, input.Catatan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "surat berhasil didistribusikan ke jabatan terkait"})
}

func (ctrl *SuratMasukController) InboxDistribusi(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, total, err := ctrl.distribusiSvc.GetInbox(userID, nil, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memuat inbox distribusi"})
		return
	}

	items := make([]gin.H, 0, len(data))
	for _, d := range data {
		items = append(items, gin.H{
			"id":             d.ID,
			"id_surat_masuk": d.IDSuratMasuk,
			"no_surat":       d.SuratMasuk.NoSurat,
			"perihal":        d.SuratMasuk.PerihalSurat,
			"asal_surat":     d.SuratMasuk.AsalSurat,
			"jabatan":        d.Jabatan.NamaJabatan,
			"catatan":        d.Catatan,
			"status":         d.Status,
			"distribute_at":  d.DistributeAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
		"page":  page,
	})
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
