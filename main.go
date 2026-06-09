package main

import (
	"log"

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

	// REPOSITORIES
	disposisiRepo := repositories.NewDisposisiRepository(db)
	suratMasukRepo := repositories.NewSuratMasukRepository(db)
	suratKeluarRepo := repositories.NewSuratKeluarRepository(db)

	// SERVICES
	notificationSvc := services.NewNotificationService(db)
	suratKeluarSvc := services.NewSuratKeluarService(suratKeluarRepo, db)
	suratMasukSvc := services.NewSuratMasukService(suratMasukRepo, disposisiRepo, db)
	disposisiSvc := services.NewDisposisiService(disposisiRepo, suratMasukRepo, notificationSvc, db)

	// CONTROLLERS
	routes.SuratMasukCtrl = controllers.NewSuratMasukController(suratMasukSvc, disposisiSvc)
	routes.SuratKeluarCtrl = controllers.NewSuratKeluarController(suratKeluarSvc)
	routes.DisposisiCtrl = controllers.NewDisposisiController(disposisiSvc)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://127.0.0.1:5500",
			"http://localhost:5500",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r)

	r.Run(":7000")
}
