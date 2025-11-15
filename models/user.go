package models

import (
    "time"

    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Name      string         `json:"name"`
    Email     string         `gorm:"unique" json:"email"`
    Password  string         `json:"-"`
    Role      string         `gorm:"default:user" json:"role"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
