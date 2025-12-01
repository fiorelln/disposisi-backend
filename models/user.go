package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Name      string         `json:"name"`
    Email     string         `gorm:"unique;not null" json:"email"`
    Password  string         `json:"-"` 
    Role      string         `gorm:"type:user_role;default:'guru'" json:"role"`
    Status    string         `gorm:"type:varchar(20);default:'pending'" json:"status"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
