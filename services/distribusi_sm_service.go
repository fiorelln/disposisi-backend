package services

import (
	"errors"
	"time"

	"github.com/fiorelln/disposisi/models"
	"github.com/fiorelln/disposisi/repositories"
	"gorm.io/gorm"
)

type DistribusiSMService interface {
	Distribute(suratMasukID uint, jabatanIDs []uint, catatan string) error
	GetInbox(userID uint, jabatanIDs []uint, page, pageSize int) ([]models.DistribusiSM, int64, error)
	GetBySuratMasukID(suratMasukID uint) ([]models.DistribusiSM, error)
	MarkAsRead(distribusiID uint) error
	CountUnread(jabatanIDs []uint) int64
}

type distribusiSMServiceImpl struct {
	repo      repositories.DistribusiSMRepository
	suratRepo repositories.SuratMasukRepository
	db        *gorm.DB
	notifSvc  NotificationService
}

func NewDistribusiSMService(
	repo repositories.DistribusiSMRepository,
	suratRepo repositories.SuratMasukRepository,
	db *gorm.DB,
	notifSvc NotificationService,
) DistribusiSMService {
	return &distribusiSMServiceImpl{
		repo:      repo,
		suratRepo: suratRepo,
		db:        db,
		notifSvc:  notifSvc,
	}
}

func (s *distribusiSMServiceImpl) Distribute(suratMasukID uint, jabatanIDs []uint, catatan string) error {
	surat, err := s.suratRepo.GetByID(suratMasukID)
	if err != nil {
		return err
	}

	if surat.StatusAlur != "diteruskan" {
		return errors.New("surat harus dikembalikan dari kepala sekolah sebelum didistribusikan")
	}

	now := time.Now()
	records := make([]models.DistribusiSM, 0, len(jabatanIDs))
	for _, jid := range jabatanIDs {
		records = append(records, models.DistribusiSM{
			IDSuratMasuk: suratMasukID,
			IDJabatan:    jid,
			Catatan:      catatan,
			Status:       "belum_dibaca",
			DistributeAt: now,
		})
	}

	if err := s.repo.CreateBatch(records); err != nil {
		return err
	}

	return s.suratRepo.UpdateStatusAlur(suratMasukID, "selesai")
}

func (s *distribusiSMServiceImpl) GetInbox(userID uint, jabatanIDs []uint, page, pageSize int) ([]models.DistribusiSM, int64, error) {
	if len(jabatanIDs) == 0 {
		var ujs []models.UserJabatan
		if err := s.db.Where("id_user = ?", userID).Find(&ujs).Error; err != nil {
			return nil, 0, err
		}
		for _, uj := range ujs {
			jabatanIDs = append(jabatanIDs, uj.JabatanID)
		}
	}
	return s.repo.GetByUserJabatanIDs(jabatanIDs, page, pageSize)
}

func (s *distribusiSMServiceImpl) GetBySuratMasukID(suratMasukID uint) ([]models.DistribusiSM, error) {
	return s.repo.GetBySuratMasukID(suratMasukID)
}

func (s *distribusiSMServiceImpl) MarkAsRead(distribusiID uint) error {
	return s.repo.MarkAsRead(distribusiID)
}

func (s *distribusiSMServiceImpl) CountUnread(jabatanIDs []uint) int64 {
	return s.repo.CountUnreadByJabatanIDs(jabatanIDs)
}
