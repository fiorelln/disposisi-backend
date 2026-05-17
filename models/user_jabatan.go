package models

type UserJabatan struct {
	UserID    uint     `gorm:"column:id_user" json:"user_id"`
	JabatanID uint     `gorm:"column:id_jabatan" json:"jabatan_id"`
	IsPrimary bool     `gorm:"column:is_primary" json:"is_primary"`

	Jabatan   Jabatan  `gorm:"foreignKey:JabatanID" json:"jabatan"`
}

func (UserJabatan) TableName() string {
	return "user_jabatan"
}
