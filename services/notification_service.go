package services

import (
	"fmt"
	"time"

	"github.com/fiorelln/disposisi/models"
	"gorm.io/gorm"
)

// NotificationService interface
type NotificationService interface {
	NotifyDisposisiReceived(disposisiID uint, recipientID uint) error
	NotifyDisposisiCompleted(disposisiID uint, senderID uint) error
	NotifyDisposisiRejected(disposisiID uint, senderID uint) error
	NotifyDisposisiForwarded(disposisiID uint, newRecipientID uint) error
}

// NotificationServiceImpl - service implementation
type NotificationServiceImpl struct {
	db *gorm.DB
}

// NewNotificationService - create new notification service
func NewNotificationService(db *gorm.DB) NotificationService {
	return &NotificationServiceImpl{db: db}
}

// ===== NOTIFICATION OPERATIONS =====

// NotifyDisposisiReceived - notify ketika menerima disposisi baru
func (s *NotificationServiceImpl) NotifyDisposisiReceived(disposisiID uint, recipientID uint) error {
	// Get disposisi detail
	var disposisi models.Disposisi
	if err := s.db.
		Preload("FromUser").
		Preload("SuratMasuk").
		First(&disposisi, disposisiID).Error; err != nil {
		return fmt.Errorf("failed to get disposisi: %w", err)
	}

	// Create notifikasi record
	notifikasi := &models.Notifikasi{
		IDPenerima:     recipientID,
		Jenis:          "disposisi_received",
		Judul:          "Disposisi Baru Diterima",
		Pesan:          fmt.Sprintf("Anda menerima disposisi dari %s tentang surat %s", disposisi.FromUser.Name, disposisi.SuratMasuk.NoSurat),
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := s.db.Create(notifikasi).Error; err != nil {
		fmt.Printf("Warning: failed to create notification: %v\n", err)
		// Don't return error, just log it
		return nil
	}

	// TODO: Implement real notification (email, push, etc)
	// s.sendEmailNotification(recipientID, notifikasi.Message)
	// s.sendPushNotification(recipientID, notifikasi.Title, notifikasi.Message)

	return nil
}

// NotifyDisposisiCompleted - notify ketika disposisi completed
func (s *NotificationServiceImpl) NotifyDisposisiCompleted(disposisiID uint, senderID uint) error {
	var disposisi models.Disposisi
	if err := s.db.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		First(&disposisi, disposisiID).Error; err != nil {
		return fmt.Errorf("failed to get disposisi: %w", err)
	}

	message := fmt.Sprintf("Disposisi dari Anda kepada %s tentang surat %s telah selesai diproses", disposisi.ToUser.Name, disposisi.SuratMasuk.NoSurat)

	notifikasi := &models.Notifikasi{
		IDPenerima:     senderID,
		Jenis:          "disposisi_completed",
		Judul:          "Disposisi Selesai",
		Pesan:          message,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := s.db.Create(notifikasi).Error; err != nil {
		fmt.Printf("Warning: failed to create notification: %v\n", err)
		return nil
	}

	return nil
}

// NotifyDisposisiRejected - notify ketika disposisi ditolak
func (s *NotificationServiceImpl) NotifyDisposisiRejected(disposisiID uint, senderID uint) error {
	var disposisi models.Disposisi
	if err := s.db.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		First(&disposisi, disposisiID).Error; err != nil {
		return fmt.Errorf("failed to get disposisi: %w", err)
	}

	message := fmt.Sprintf("Disposisi dari Anda kepada %s tentang surat %s ditolak", disposisi.ToUser.Name, disposisi.SuratMasuk.NoSurat)

	notifikasi := &models.Notifikasi{
		IDPenerima:     senderID,
		Jenis:          "disposisi_rejected",
		Judul:          "Disposisi Ditolak",
		Pesan:          message,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := s.db.Create(notifikasi).Error; err != nil {
		fmt.Printf("Warning: failed to create notification: %v\n", err)
		return nil
	}

	return nil
}

// NotifyDisposisiForwarded - notify ketika disposisi di-forward lagi
func (s *NotificationServiceImpl) NotifyDisposisiForwarded(disposisiID uint, newRecipientID uint) error {
	var disposisi models.Disposisi
	if err := s.db.
		Preload("FromUser").
		Preload("ToUser").
		Preload("SuratMasuk").
		First(&disposisi, disposisiID).Error; err != nil {
		return fmt.Errorf("failed to get disposisi: %w", err)
	}

	message := fmt.Sprintf("Disposisi tentang surat %s di-forward dari %s", disposisi.SuratMasuk.NoSurat, disposisi.FromUser.Name)

	notifikasi := &models.Notifikasi{
		IDPenerima:     newRecipientID,
		Jenis:          "disposisi_forwarded",
		Judul:          "Disposisi Di-forward",
		Pesan:          message,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := s.db.Create(notifikasi).Error; err != nil {
		fmt.Printf("Warning: failed to create notification: %v\n", err)
		return nil
	}

	return nil
}
