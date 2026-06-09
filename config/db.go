package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fiorelln/disposisi/models"
	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var JwtKey []byte

func ConnectDB() {
	_ = godotenv.Load()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	JwtKey = []byte(secret)

	databaseURL := os.Getenv("DATABASE_URL")

	var dsn string
	if databaseURL != "" {
		dsn = databaseURL
	} else {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}

		if password != "" {
			dsn = fmt.Sprintf(
				"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
				host, user, password, dbname, port, sslmode,
			)
		} else {
			dsn = fmt.Sprintf(
				"host=%s user=%s dbname=%s port=%s sslmode=%s",
				host, user, dbname, port, sslmode,
			)
		}
	}

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	sqlDB, err := DB.DB()

	if err != nil {
		log.Fatal("Gagal ambil sql.DB:", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database berhasil terhubung")

	err = DB.AutoMigrate(
		&models.User{},
		&models.Jabatan{},
		&models.UserJabatan{},
		&models.OTP{},
		&models.SuratMasuk{},
		&models.SuratKeluar{},
		&models.Disposisi{},
		&models.Notifikasi{},
		&models.Log{},
	)
	if err != nil {
		log.Fatal("Gagal migrasi database:", err)
	}

	log.Println("Migrasi database berhasil")
}
