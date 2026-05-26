package main

import (
	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/controllers"
	"github.com/fiorelln/disposisi/repositories"
	"github.com/fiorelln/disposisi/services"
	"github.com/gin-gonic/gin"
)

// InitializeDependencies - Setup semua dependencies dengan Dependency Injection
// Panggil ini di main() atau di routes setup
func InitializeDependencies(r *gin.Engine) {
	db := config.DB

	// ===== REPOSITORIES =====
	disposisiRepo := repositories.NewDisposisiRepository(db)

	// ===== SERVICES =====
	permissionSvc := services.NewPermissionService(db)
	notificationSvc := services.NewNotificationService(db)
	disposisiSvc := services.NewDisposisiService(
		disposisiRepo,
		permissionSvc,
		notificationSvc,
		db,
	)

	// ===== CONTROLLERS =====
	disposisiCtrl := controllers.NewDisposisiController(disposisiSvc)

	// ===== ROUTES =====
	// Disposisi endpoints
	disposisiRoutes := r.Group("/disposisi")
	{
		// Forwarding
		disposisiRoutes.POST("/:id/forward", disposisiCtrl.ForwardDisposisi)

		// Completion
		disposisiRoutes.POST("/:id/complete", disposisiCtrl.CompleteDisposisi)
		disposisiRoutes.POST("/:id/reject", disposisiCtrl.RejectDisposisi)

		// Read status
		disposisiRoutes.POST("/:id/read", disposisiCtrl.MarkAsRead)
		disposisiRoutes.POST("/read-batch", disposisiCtrl.MarkAsReadBatch)

		// Queries
		disposisiRoutes.GET("/inbox", disposisiCtrl.GetInbox)
		disposisiRoutes.GET("/sent", disposisiCtrl.GetSentItems)
		disposisiRoutes.GET("/:id", disposisiCtrl.GetDisposisiDetail)
		disposisiRoutes.GET("/stats", disposisiCtrl.GetStats)
	}

	// Surat endpoints
	suratRoutes := r.Group("/surat")
	{
		suratRoutes.GET("/:surat_id/history", disposisiCtrl.GetHistory)
	}
}

// ===== ALTERNATIVE: Manual Dependency Injection (jika prefer explicit) =====

/*
Contoh setup manual untuk debugging atau testing:

func setupDisposisiRoutes(r *gin.Engine) {
	// Get database instance
	db := config.DB
	
	// Create repository
	disposisiRepo := repositories.NewDisposisiRepository(db)
	
	// Create services dengan explicit dependencies
	permissionSvc := services.NewPermissionService(db)
	notificationSvc := services.NewNotificationService(db)
	disposisiSvc := services.NewDisposisiService(
		disposisiRepo,
		permissionSvc,
		notificationSvc,
		db,
	)
	
	// Create controller dengan service dependency
	disposisiCtrl := controllers.NewDisposisiController(disposisiSvc)
	
	// Register routes
	r.POST("/disposisi/:id/forward", disposisiCtrl.ForwardDisposisi)
	r.POST("/disposisi/:id/complete", disposisiCtrl.CompleteDisposisi)
	r.GET("/disposisi/inbox", disposisiCtrl.GetInbox)
	r.GET("/disposisi/sent", disposisiCtrl.GetSentItems)
	r.GET("/surat/:surat_id/history", disposisiCtrl.GetHistory)
}

// Panggil di main():
func main() {
	config.ConnectDB()
	
	r := gin.Default()
	
	// Setup CORS, middlewares, etc
	
	// Initialize disposisi routes
	setupDisposisiRoutes(r)
	
	// Start server
	r.Run(":7000")
}
*/
