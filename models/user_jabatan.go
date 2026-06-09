package models

type UserJabatan struct {
	UserID    uint `gorm:"column:id_user;primaryKey" json:"user_id"`
	JabatanID uint `gorm:"column:id_jabatan;primaryKey" json:"jabatan_id"`
	IsPrimary bool `gorm:"column:is_primary;default:false" json:"is_primary"`

	User    User    `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Jabatan Jabatan `gorm:"foreignKey:JabatanID;references:ID" json:"jabatan"`
}

func (UserJabatan) TableName() string {
	return "user_jabatan"
}
