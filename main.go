package main

import (
	"log"
	"os"
	"strings"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/controllers"
	"github.com/fiorelln/disposisi/repositories"
	"github.com/fiorelln/disposisi/routes"
	"github.com/fiorelln/disposisi/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Gagal load .env")
	}

	config.ConnectDB()
	db := config.DB

	disposisiRepo := repositories.NewDisposisiRepository(db)
	suratMasukRepo := repositories.NewSuratMasukRepository(db)
	suratKeluarRepo := repositories.NewSuratKeluarRepository(db)

	notificationSvc := services.NewNotificationService(db)
	suratKeluarSvc := services.NewSuratKeluarService(suratKeluarRepo, db)
	suratMasukSvc := services.NewSuratMasukService(suratMasukRepo, disposisiRepo, notificationSvc, db)
	disposisiSvc := services.NewDisposisiService(disposisiRepo, suratMasukRepo, notificationSvc, db)

	routes.SuratMasukCtrl = controllers.NewSuratMasukController(suratMasukSvc, disposisiSvc)
	routes.SuratKeluarCtrl = controllers.NewSuratKeluarController(suratKeluarSvc)
	routes.DisposisiCtrl = controllers.NewDisposisiController(disposisiSvc)

	os.MkdirAll("uploads/surat_masuk", os.ModePerm)
	os.MkdirAll("uploads/surat_keluar", os.ModePerm)

	r := gin.Default()

	origins := os.Getenv("CORS_ALLOW_ORIGINS")
	if origins == "" {
		origins = "*"
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(origins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: origins != "*",
	}))

	routes.SetupRoutes(r)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "7000"
	}
	r.Run(":" + port)
}
