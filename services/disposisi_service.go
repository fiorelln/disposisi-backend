package services

import (
	"errors"

	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/repositories"
	"gorm.io/gorm"
)

type DisposisiService interface {
	GetInbox(userID uint, page, pageSize int) ([]models.Disposisi, int64, error)
	GetSentItems(userID uint, page, pageSize int) ([]models.Disposisi, int64, error)
	GetBySuratMasukID(suratMasukID uint) ([]models.Disposisi, error)
	GetDisposisiByID(id uint) (*models.Disposisi, error)
	MarkAsRead(disposisiID uint, userID uint) error
	WakaAction(disposisiID uint, userID uint, isiDisposisi string, batasWaktu string, tanggapanSaran string, prosesLanjut string, koordinasiKonfirmasi string, penerimaID *uint, jabatanPenerimaID *uint) error
	CompleteDisposisi(disposisiID uint, userID uint) error
	GetStats(userID uint) (map[string]interface{}, error)
}

type DisposisiServiceImpl struct {
	repo            repositories.DisposisiRepository
	suratRepo       repositories.SuratMasukRepository
	notificationSvc NotificationService
	db              *gorm.DB
}

func NewDisposisiService(
	repo repositories.DisposisiRepository,
	suratRepo repositories.SuratMasukRepository,
	notificationSvc NotificationService,
	db *gorm.DB,
) DisposisiService {
	return &DisposisiServiceImpl{
		repo:            repo,
		suratRepo:       suratRepo,
		notificationSvc: notificationSvc,
		db:              db,
	}
}

func (s *DisposisiServiceImpl) GetInbox(userID uint, page, pageSize int) ([]models.Disposisi, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var disposisi []models.Disposisi
	var total int64

	offset := (page - 1) * pageSize

	countErr := s.db.
		Where("id_penerima = ?", userID).
		Model(&models.Disposisi{}).
		Count(&total).Error
	if countErr != nil {
		return nil, 0, countErr
	}

	err := s.db.
		Where("id_penerima = ?", userID).
		Preload("SuratMasuk").
		Preload("Penerima").
		Order("tanggal_disposisi DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&disposisi).Error

	return disposisi, total, err
}

func (s *DisposisiServiceImpl) GetSentItems(userID uint, page, pageSize int) ([]models.Disposisi, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var disposisi []models.Disposisi
	var total int64

	offset := (page - 1) * pageSize

	countErr := s.db.
		Where("id_kepsek = ?", userID).
		Or("EXISTS (SELECT 1 FROM surat_masuk WHERE surat_masuk.id_surat_masuk = disposisi.id_surat_masuk AND surat_masuk.user_verifikasi = ?)", userID).
		Model(&models.Disposisi{}).
		Count(&total).Error
	if countErr != nil {
		return nil, 0, countErr
	}

	err := s.db.
		Where("id_kepsek = ?", userID).
		Or("EXISTS (SELECT 1 FROM surat_masuk WHERE surat_masuk.id_surat_masuk = disposisi.id_surat_masuk AND surat_masuk.user_verifikasi = ?)", userID).
		Preload("SuratMasuk").
		Preload("Penerima").
		Order("tanggal_disposisi DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&disposisi).Error

	return disposisi, total, err
}

func (s *DisposisiServiceImpl) GetBySuratMasukID(suratMasukID uint) ([]models.Disposisi, error) {
	var disposisi []models.Disposisi
	err := s.db.
		Where("id_surat_masuk = ?", suratMasukID).
		Preload("SuratMasuk").
		Preload("Kepsek").
		Preload("Penerima").
		Preload("JabatanPenerima").
		Order("tanggal_disposisi ASC").
		Find(&disposisi).Error
	return disposisi, err
}

func (s *DisposisiServiceImpl) GetDisposisiByID(id uint) (*models.Disposisi, error) {
	var disposisi models.Disposisi
	err := s.db.
		Preload("SuratMasuk").
		Preload("Kepsek").
		Preload("Penerima").
		Preload("JabatanPenerima").
		First(&disposisi, id).Error
	if err != nil {
		return nil, errors.New("disposisi tidak ditemukan")
	}
	return &disposisi, nil
}

func (s *DisposisiServiceImpl) MarkAsRead(disposisiID uint, userID uint) error {
	var disposisi models.Disposisi
	if err := s.db.First(&disposisi, disposisiID).Error; err != nil {
		return errors.New("disposisi tidak ditemukan")
	}

	if disposisi.PenerimaID == nil || *disposisi.PenerimaID != userID {
		return errors.New("anda bukan penerima disposisi ini")
	}

	return s.db.Model(&models.Disposisi{}).
		Where("id_disposisi = ?", disposisiID).
		Update("status_disposisi", "dibaca").Error
}

func (s *DisposisiServiceImpl) WakaAction(
	disposisiID uint,
	userID uint,
	isiDisposisi string,
	batasWaktu string,
	tanggapanSaran string,
	prosesLanjut string,
	koordinasiKonfirmasi string,
	penerimaID *uint,
	jabatanPenerimaID *uint,
) error {
	var disposisi models.Disposisi
	if err := s.db.First(&disposisi, disposisiID).Error; err != nil {
		return errors.New("disposisi tidak ditemukan")
	}

	if disposisi.PenerimaID == nil || *disposisi.PenerimaID != userID {
		return errors.New("anda bukan penerima disposisi ini")
	}

	tx := s.db.Begin()

	updates := map[string]interface{}{
		"status_disposisi":       "sedang_dikerjakan",
		"isi_disposisi":          isiDisposisi,
		"batas_waktu":            batasWaktu,
		"tanggapan_saran":        tanggapanSaran,
		"proses_lanjut":          prosesLanjut,
		"koordinasi_konfirmasi":  koordinasiKonfirmasi,
	}

	if err := tx.Model(&models.Disposisi{}).
		Where("id_disposisi = ?", disposisiID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if penerimaID != nil && *penerimaID > 0 {
		childDisposisi := &models.Disposisi{
			SuratMasukID:      disposisi.SuratMasukID,
			PenerimaID:        penerimaID,
			JabatanPenerimaID: jabatanPenerimaID,
			StatusDisposisi:   "belum_dibaca",
		}

		if err := tx.Create(childDisposisi).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func (s *DisposisiServiceImpl) CompleteDisposisi(disposisiID uint, userID uint) error {
	var disposisi models.Disposisi
	if err := s.db.First(&disposisi, disposisiID).Error; err != nil {
		return errors.New("disposisi tidak ditemukan")
	}

	if disposisi.PenerimaID == nil || *disposisi.PenerimaID != userID {
		return errors.New("anda bukan penerima disposisi ini")
	}

	tx := s.db.Begin()

	if err := tx.Model(&models.Disposisi{}).
		Where("id_disposisi = ?", disposisiID).
		Updates(map[string]interface{}{
			"status_disposisi": "selesai",
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.SuratMasuk{}).
		Where("id_surat_masuk = ?", disposisi.SuratMasukID).
		Update("status_alur", "selesai").Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	s.notificationSvc.NotifyDisposisiCompleted(disposisiID, 0)

	return nil
}

func (s *DisposisiServiceImpl) GetStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalDisposisi int64
	s.db.Model(&models.Disposisi{}).
		Where("id_penerima = ?", userID).
		Count(&totalDisposisi)
	stats["total_disposisi"] = totalDisposisi

	var belumDibaca int64
	s.db.Model(&models.Disposisi{}).
		Where("id_penerima = ? AND status_disposisi = ?", userID, "belum_dibaca").
		Count(&belumDibaca)
	stats["belum_dibaca"] = belumDibaca

	var sedangDikerjakan int64
	s.db.Model(&models.Disposisi{}).
		Where("id_penerima = ? AND status_disposisi = ?", userID, "sedang_dikerjakan").
		Count(&sedangDikerjakan)
	stats["sedang_dikerjakan"] = sedangDikerjakan

	var selesai int64
	s.db.Model(&models.Disposisi{}).
		Where("id_penerima = ? AND status_disposisi = ?", userID, "selesai").
		Count(&selesai)
	stats["selesai"] = selesai

	return stats, nil
}
