package services

import (
	"fmt"
	"time"

	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

type NotificationService interface {
	NotifyDisposisiReceived(disposisiID uint, recipientID uint) error
	NotifyDisposisiCompleted(disposisiID uint, senderID uint) error
	NotifyDisposisiRejected(disposisiID uint, senderID uint) error
}

type NotificationServiceImpl struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) NotificationService {
	return &NotificationServiceImpl{db: db}
}

func (s *NotificationServiceImpl) NotifyDisposisiReceived(disposisiID uint, recipientID uint) error {
	var disposisi models.Disposisi
	if err := s.db.
		Preload("SuratMasuk").
		First(&disposisi, disposisiID).Error; err != nil {
		return fmt.Errorf("failed to get disposisi: %w", err)
	}

	notifikasi := &models.Notifikasi{
		IDPenerima: recipientID,
		Jenis:      "disposisi_received",
		Judul:      "Disposisi Baru Diterima",
		Pesan: fmt.Sprintf(
			"Anda menerima disposisi tentang surat %s",
			disposisi.SuratMasuk.NoSurat,
		),
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(notifikasi).Error; err != nil {
		fmt.Printf("Warning: failed to create notification: %v\n", err)
		return nil
	}

	return nil
}

func (s *NotificationServiceImpl) NotifyDisposisiCompleted(disposisiID uint, senderID uint) error {
	var disposisi models.Disposisi
	if err := s.db.
		Preload("SuratMasuk").
		First(&disposisi, disposisiID).Error; err != nil {
		return fmt.Errorf("failed to get disposisi: %w", err)
	}

	message := fmt.Sprintf(
		"Disposisi tentang surat %s telah selesai diproses",
		disposisi.SuratMasuk.NoSurat,
	)

	notifikasi := &models.Notifikasi{
		IDPenerima: senderID,
		Jenis:      "disposisi_completed",
		Judul:      "Disposisi Selesai",
		Pesan:      message,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(notifikasi).Error; err != nil {
		fmt.Printf("Warning: failed to create notification: %v\n", err)
		return nil
	}

	return nil
}

func (s *NotificationServiceImpl) NotifyDisposisiRejected(disposisiID uint, senderID uint) error {
	var disposisi models.Disposisi
	if err := s.db.
		Preload("SuratMasuk").
		First(&disposisi, disposisiID).Error; err != nil {
		return fmt.Errorf("failed to get disposisi: %w", err)
	}

	message := fmt.Sprintf(
		"Disposisi tentang surat %s ditolak",
		disposisi.SuratMasuk.NoSurat,
	)

	notifikasi := &models.Notifikasi{
		IDPenerima: senderID,
		Jenis:      "disposisi_rejected",
		Judul:      "Disposisi Ditolak",
		Pesan:      message,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(notifikasi).Error; err != nil {
		fmt.Printf("Warning: failed to create notification: %v\n", err)
		return nil
	}

	return nil
}
