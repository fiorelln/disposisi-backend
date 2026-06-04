package services

import (
	"errors"
	"time"

	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/repositories"
)

type SuratMasukService interface {
	Register(noSurat, perihal, asal string, filePDF string) (*models.SuratMasuk, error)
	ForwardToPrincipal(suratID uint) error
	Review(suratID uint, kepsekID uint, approvalStatus string, catatan string, tanggapan string, proses string, koordinasi string) error
	GetByID(id uint) (*models.SuratMasuk, error)
	List(page, pageSize int, status string) ([]models.SuratMasuk, int64, error)
}

type suratMasukService struct {
	suratRepo     repositories.SuratMasukRepository
	disposisiRepo repositories.DisposisiRepository
}

func NewSuratMasukService(
	suratRepo repositories.SuratMasukRepository,
	disposisiRepo repositories.DisposisiRepository,
) SuratMasukService {
	return &suratMasukService{
		suratRepo:     suratRepo,
		disposisiRepo: disposisiRepo,
	}
}

func (s *suratMasukService) Register(noSurat, perihal, asal string, filePDF string) (*models.SuratMasuk, error) {
	now := time.Now()
	surat := &models.SuratMasuk{
		NoSurat:         noSurat,
		PerihalSurat:    perihal,
		AsalSurat:       asal,
		FilePDF:         filePDF,
		TanggalSurat:    &now,
		TanggalDiterima: &now,
		StatusAlur:      "diterima_tu",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.suratRepo.Create(surat); err != nil {
		return nil, err
	}
	return surat, nil
}

func (s *suratMasukService) ForwardToPrincipal(suratID uint) error {
	surat, err := s.suratRepo.GetByID(suratID)
	if err != nil {
		return err
	}

	if surat.StatusAlur != "diterima_tu" {
		return errors.New("surat tidak dapat diteruskan ke kepala sekolah pada tahap ini")
	}

	principalID := uint(1)

	disposisi := &models.Disposisi{
		SuratMasukID: suratID,
		FromUserID:   1,
		ToUserID:     principalID,
		Status:       models.StatusPending,
		Sifat:        "segera",
		Dibaca:       false,
		Catatan:      "",
	}

	if err := s.disposisiRepo.Create(disposisi); err != nil {
		return err
	}

	if err := s.suratRepo.SetDisposisiAktif(suratID, disposisi.ID); err != nil {
		return err
	}

	return s.suratRepo.UpdateStatusAlur(suratID, "disposisi_kepsek")
}

func (s *suratMasukService) Review(suratID uint, kepsekID uint, approvalStatus string, catatan string, tanggapan string, proses string, koordinasi string) error {
	surat, err := s.suratRepo.GetByID(suratID)
	if err != nil {
		return err
	}

	if surat.IDDisposisiAktif == nil {
		return errors.New("tidak ditemukan record disposisi aktif untuk surat ini")
	}

	disposisi, err := s.disposisiRepo.GetByID(*surat.IDDisposisiAktif)
	if err != nil {
		return err
	}

	disposisi.TanggapanSaran = tanggapan
	disposisi.ProsesLanjut = proses
	disposisi.KoordinasiKonfirmasi = koordinasi
	disposisi.Catatan = catatan
	if approvalStatus == "disetujui" {
		disposisi.Status = models.StatusCompleted
	} else {
		disposisi.Status = models.StatusRejected
	}
	disposisi.FromUserID = kepsekID
	now := time.Now()
	disposisi.CompleteAt = &now
	disposisi.Dibaca = true

	if err := s.disposisiRepo.Update(disposisi); err != nil {
		return err
	}

	return s.suratRepo.UpdateStatusAlur(suratID, "diteruskan")
}



func (s *suratMasukService) GetByID(id uint) (*models.SuratMasuk, error) {
	return s.suratRepo.GetByID(id)
}

func (s *suratMasukService) List(page, pageSize int, status string) ([]models.SuratMasuk, int64, error) {
	return s.suratRepo.ListAll(page, pageSize, status)
}
