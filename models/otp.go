package models

import "time"

type OTP struct {
	ID        uint      `gorm:"column:id_otp;primaryKey"`
	UserID    uint      `gorm:"column:id_user"`
	KodeOTP   string    `gorm:"column:kode_otp"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
	IsUsed    bool      `gorm:"column:is_used"`
}

func (OTP) TableName() string {
	return "otp"
}