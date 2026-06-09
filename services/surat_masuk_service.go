package services

import (
	"errors"
	"time"

	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/repositories"
	"gorm.io/gorm"
)

type SuratMasukService interface {
	Register(noSurat, perihal, asal string, filePDF string) (*models.SuratMasuk, error)
	ForwardToPrincipal(suratID uint, tuUserID uint) error
	Review(suratID uint, kepsekID uint, statusApproval string, catatan string) error
	DistributeToUser(suratID uint, tuUserID uint, penerimaID uint, jabatanPenerimaID uint, catatan string) (*models.Disposisi, error)
	GetByID(id uint) (*models.SuratMasuk, error)
	List(page, pageSize int, status string) ([]models.SuratMasuk, int64, error)
}

var wakaJabatan = map[string]bool{
	"waka kesiswaan": true,
	"waka kurikulum": true,
	"waka sarpras":   true,
	"waka humas":     true,
}

type suratMasukService struct {
	suratRepo       repositories.SuratMasukRepository
	disposisiRepo   repositories.DisposisiRepository
	notificationSvc NotificationService
	db              *gorm.DB
}

func NewSuratMasukService(
	suratRepo repositories.SuratMasukRepository,
	disposisiRepo repositories.DisposisiRepository,
	notificationSvc NotificationService,
	db *gorm.DB,
) SuratMasukService {
	return &suratMasukService{
		suratRepo:       suratRepo,
		disposisiRepo:   disposisiRepo,
		notificationSvc: notificationSvc,
		db:              db,
	}
}

func (s *suratMasukService) Register(noSurat, perihal, asal string, filePDF string) (*models.SuratMasuk, error) {
	now := time.Now()
	surat := &models.SuratMasuk{
		NoSurat:          noSurat,
		PerihalSurat:     perihal,
		AsalSurat:        asal,
		FilePDF:          filePDF,
		TanggalSurat:     &now,
		TanggalDiterima:  &now,
		StatusVerifikasi: "menunggu",
		StatusAlur:       "diterima_tu",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.suratRepo.Create(surat); err != nil {
		return nil, err
	}
	return surat, nil
}

func (s *suratMasukService) ForwardToPrincipal(suratID uint, tuUserID uint) error {
	surat, err := s.suratRepo.GetByID(suratID)
	if err != nil {
		return err
	}

	if surat.StatusAlur != "diterima_tu" {
		return errors.New("surat tidak dapat diteruskan ke Kepala Sekolah pada tahap ini")
	}

	var principal models.User
	if err := s.db.Joins("JOIN user_jabatan ON user_jabatan.id_user = users.id").
		Joins("JOIN jabatan ON jabatan.id_jabatan = user_jabatan.id_jabatan").
		Where("jabatan.nama_jabatan = ?", "kepala sekolah").
		First(&principal).Error; err != nil {
		return errors.New("tidak ditemukan user dengan jabatan Kepala Sekolah")
	}

	disposisi := &models.Disposisi{
		SuratMasukID:    suratID,
		KepsekID:        &principal.ID,
		StatusDisposisi: "belum_dibaca",
		StatusApproval:  "menunggu",
	}

	if err := s.disposisiRepo.Create(disposisi); err != nil {
		return err
	}

	return s.suratRepo.UpdateStatusAlur(suratID, "disposisi_kepsek")
}

func (s *suratMasukService) Review(suratID uint, kepsekID uint, statusApproval string, catatan string) error {
	surat, err := s.suratRepo.GetByID(suratID)
	if err != nil {
		return err
	}

	if surat.StatusAlur != "disposisi_kepsek" {
		return errors.New("surat tidak dalam tahap disposisi Kepala Sekolah")
	}

	var disposisi models.Disposisi
	if err := s.db.Where("id_surat_masuk = ? AND id_kepsek = ? AND status_approval = ?",
		suratID, kepsekID, "menunggu").
		First(&disposisi).Error; err != nil {
		return errors.New("tidak ditemukan disposisi yang menunggu persetujuan anda")
	}

	now := time.Now()

	if statusApproval == "disetujui" {
		updates := map[string]interface{}{
			"status_approval": "disetujui",
			"approval_at":     now,
			"catatan_kepsek":  catatan,
		}
		s.db.Model(&disposisi).Updates(updates)

		s.suratRepo.Verify(suratID, kepsekID, "disetujui", catatan)
		return s.suratRepo.UpdateStatusAlur(suratID, "disetujui_kembali_ke_tu")
	}

	updates := map[string]interface{}{
		"status_approval":  "ditolak",
		"approval_at":      now,
		"catatan_kepsek":   catatan,
		"status_disposisi": "selesai",
	}
	s.db.Model(&disposisi).Updates(updates)

	s.suratRepo.Verify(suratID, kepsekID, "ditolak", catatan)
	return s.suratRepo.UpdateStatusAlur(suratID, "selesai")
}

func (s *suratMasukService) DistributeToUser(suratID uint, tuUserID uint, penerimaID uint, jabatanPenerimaID uint, catatan string) (*models.Disposisi, error) {
	surat, err := s.suratRepo.GetByID(suratID)
	if err != nil {
		return nil, err
	}

	if surat.StatusAlur != "disetujui_kembali_ke_tu" {
		return nil, errors.New("surat harus disetujui Kepala Sekolah dan dikembalikan ke TU terlebih dahulu")
	}

	var jabatan models.Jabatan
	if err := s.db.First(&jabatan, jabatanPenerimaID).Error; err != nil {
		return nil, errors.New("jabatan penerima tidak ditemukan")
	}

	if !wakaJabatan[jabatan.NamaJabatan] {
		return nil, errors.New("TU hanya dapat mendistribusikan surat kepada Waka (waka kesiswaan, waka kurikulum, waka sarpras, waka humas)")
	}

	var penerima models.User
	if err := s.db.First(&penerima, penerimaID).Error; err != nil {
		return nil, errors.New("penerima tidak ditemukan")
	}

	disposisi := &models.Disposisi{
		SuratMasukID:      suratID,
		PenerimaID:        &penerimaID,
		JabatanPenerimaID: &jabatanPenerimaID,
		StatusDisposisi:   "belum_dibaca",
		CatatanKepsek:     catatan,
	}

	if err := s.disposisiRepo.Create(disposisi); err != nil {
		return nil, err
	}

	if err := s.suratRepo.UpdateStatusAlur(suratID, "diteruskan"); err != nil {
		return nil, err
	}

	go s.notificationSvc.NotifyDisposisiReceived(disposisi.ID, penerimaID)

	return disposisi, nil
}

func (s *suratMasukService) GetByID(id uint) (*models.SuratMasuk, error) {
	return s.suratRepo.GetByID(id)
}

func (s *suratMasukService) List(page, pageSize int, status string) ([]models.SuratMasuk, int64, error) {
	return s.suratRepo.ListAll(page, pageSize, status)
}
