package models

type Jabatan struct {
	ID          uint   `gorm:"primaryKey;column:id_jabatan" json:"id"`
	NamaJabatan string `gorm:"column:nama_jabatan" json:"nama_jabatan"`
	LevelAkses  string `gorm:"column:level_akses" json:"level_akses"`
}

func (Jabatan) TableName() string {
	return "jabatan"
}
