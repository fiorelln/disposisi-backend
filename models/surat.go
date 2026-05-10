package models

import "gorm.io/gorm"

type Surat struct {
	gorm.Model
	FileSurat string
	Status    string
}
