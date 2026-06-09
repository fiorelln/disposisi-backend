package controllers

import (
	"net/http"
	"strconv"

	"github.com/fiorelln/disposisi/services"
	"github.com/gin-gonic/gin"
)

type DisposisiController struct {
	service services.DisposisiService
}

func NewDisposisiController(service services.DisposisiService) *DisposisiController {
	return &DisposisiController{service: service}
}

func (ctrl *DisposisiController) GetInbox(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, total, err := ctrl.service.GetInbox(userID, page, pageSize)
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

func (ctrl *DisposisiController) GetSentItems(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	data, total, err := ctrl.service.GetSentItems(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memuat item terkirim"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *DisposisiController) GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	disposisi, err := ctrl.service.GetDisposisiByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": disposisi})
}

func (ctrl *DisposisiController) GetHistory(c *gin.Context) {
	suratIDStr := c.Param("surat_id")
	suratID, _ := strconv.ParseUint(suratIDStr, 10, 32)

	data, err := ctrl.service.GetBySuratMasukID(uint(suratID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memuat riwayat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (ctrl *DisposisiController) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	userID := c.MustGet("user_id").(uint)

	if err := ctrl.service.MarkAsRead(uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disposisi ditandai telah dibaca"})
}

func (ctrl *DisposisiController) WakaAction(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	userID := c.MustGet("user_id").(uint)

	var input struct {
		IsiDisposisi         string `json:"isi_disposisi"`
		BatasWaktu           string `json:"batas_waktu"`
		TanggapanSaran       string `json:"tanggapan_saran"`
		ProsesLanjut         string `json:"proses_lanjut"`
		KoordinasiKonfirmasi string `json:"koordinasi_konfirmasi"`
		PenerimaID           *uint  `json:"penerima_id"`
		JabatanPenerimaID    *uint  `json:"jabatan_penerima_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format data tidak valid"})
		return
	}

	if err := ctrl.service.WakaAction(uint(id), userID, input.IsiDisposisi, input.BatasWaktu, input.TanggapanSaran, input.ProsesLanjut, input.KoordinasiKonfirmasi, input.PenerimaID, input.JabatanPenerimaID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disposisi berhasil diproses"})
}

func (ctrl *DisposisiController) CompleteDisposisi(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	userID := c.MustGet("user_id").(uint)

	if err := ctrl.service.CompleteDisposisi(uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disposisi selesai"})
}

func (ctrl *DisposisiController) GetStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	stats, err := ctrl.service.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memuat statistik"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
