package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey;column:id_user" json:"id"`
	Name      string    `gorm:"column:nama" json:"name"`
	Email     string    `gorm:"column:email;unique;not null" json:"email"`
	Password  string    `gorm:"column:password" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	Jabatans []UserJabatan `gorm:"foreignKey:UserID" json:"jabatans"`
}

func (User) TableName() string {
	return "users"
}