package services

import (
	"errors"
	"time"

	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/repositories"
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
}

func NewSuratKeluarService(suratRepo repositories.SuratKeluarRepository) SuratKeluarService {
	return &suratKeluarService{suratRepo: suratRepo}
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
	return s.suratRepo.UpdateStatusAlur(suratID, "disposisi_kepsek")
}

func (s *suratKeluarService) Review(suratID uint, principalID uint, status string, notes string) error {
	if status != "disetujui" && status != "ditolak" {
		return errors.New("status persetujuan tidak valid")
	}

	err := s.suratRepo.Verify(suratID, principalID, status, notes)
	if err != nil {
		return err
	}

	if status == "disetujui" {
		return s.suratRepo.UpdateStatusAlur(suratID, "selesai")
	}
	return nil
}

func (s *suratKeluarService) GetByID(id uint) (*models.SuratKeluar, error) {
	return s.suratRepo.GetByID(id)
}

func (s *suratKeluarService) List(page, pageSize int, status string) ([]models.SuratKeluar, int64, error) {
	return s.suratRepo.ListAll(page, pageSize, status)
}
