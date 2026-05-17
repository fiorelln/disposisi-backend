package models

import "gorm.io/gorm"

type Disposisi struct {
        gorm.Model
        SuratID uint
        Tujuan  string
        Catatan string
        Status  string 
}