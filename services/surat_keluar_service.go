package services

import (
	"errors"
	"time"

	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/repositories"
	"gorm.io/gorm"
)

type SuratKeluarService interface {
	Create(noSurat, perihal, catatan, tujuan string, filePDF string) (*models.SuratKeluar, error)
	SubmitToPrincipal(suratID uint) error
	Review(suratID uint, principalID uint, status string, notes string) error
	GetByID(id uint) (*models.SuratKeluar, error)
	List(page, pageSize int, status string) ([]models.SuratKeluar, int64, error)
}

type suratKeluarService struct {
	suratRepo repositories.SuratKeluarRepository
	db        *gorm.DB
}

func NewSuratKeluarService(suratRepo repositories.SuratKeluarRepository, db *gorm.DB) SuratKeluarService {
	return &suratKeluarService{suratRepo: suratRepo, db: db}
}

func (s *suratKeluarService) Create(noSurat, perihal, catatan, tujuan string, filePDF string) (*models.SuratKeluar, error) {
	now := time.Now()
	surat := &models.SuratKeluar{
		KodeSurat:        "",
		NoSurat:          noSurat,
		Perihal:          perihal,
		Catatan:          catatan,
		Tujuan:           tujuan,
		FilePDF:          filePDF,
		TanggalSurat:     &now,
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

func (s *suratKeluarService) SubmitToPrincipal(suratID uint) error {
	surat, err := s.suratRepo.GetByID(suratID)
	if err != nil {
		return err
	}

	if surat.StatusAlur != "diterima_tu" {
		return errors.New("surat tidak dapat diajukan pada tahap ini")
	}

	return s.suratRepo.UpdateStatusAlur(suratID, "disposisi_kepsek")
}

func (s *suratKeluarService) Review(suratID uint, principalID uint, status string, notes string) error {
	if status != "disetujui" && status != "ditolak" {
		return errors.New("status persetujuan tidak valid")
	}

	surat, err := s.suratRepo.GetByID(suratID)
	if err != nil {
		return err
	}

	if surat.StatusAlur != "disposisi_kepsek" {
		return errors.New("surat tidak dalam tahap review")
	}

	err = s.suratRepo.Verify(suratID, principalID, status, notes)
	if err != nil {
		return err
	}

	if status == "disetujui" {
		return s.suratRepo.UpdateStatusAlur(suratID, "selesai")
	}

	return s.suratRepo.UpdateStatusAlur(suratID, "selesai")
}

func (s *suratKeluarService) GetByID(id uint) (*models.SuratKeluar, error) {
	return s.suratRepo.GetByID(id)
}

func (s *suratKeluarService) List(page, pageSize int, status string) ([]models.SuratKeluar, int64, error) {
	return s.suratRepo.ListAll(page, pageSize, status)
}
