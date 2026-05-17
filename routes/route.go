package routes

import (
	"github.com/fiorelln/disposisi/controllers"
	"github.com/fiorelln/disposisi/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	AuthRoutes(r)
	AdminRoutes(r)
	SuratRoutes(r)
	DisposisiRoutes(r)
	DashboardRoutes(r)
	DocsRoutes(r)
}

func AuthRoutes(r *gin.Engine) {

	auth := r.Group("/auth")

	auth.POST("/login", controllers.Login)
	auth.POST("/forgot-password", controllers.ForgotPassword)
	auth.POST("/verify-otp", controllers.VerifyOTP)
	auth.POST("/reset-password", controllers.ResetPassword)
}

func AdminRoutes(r *gin.Engine) {

	admin := r.Group("/admin")
	admin.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("Admin"))

	admin.POST("/users", controllers.Register)
}

func SuratRoutes(r *gin.Engine) {

	surat := r.Group("/surat")
	surat.Use(middlewares.AuthMiddleware())

	surat.POST("/upload", controllers.UploadSurat)
	surat.GET("/:surat_id", controllers.GetSurat)
	surat.GET("", controllers.ListSurat)
}

func DisposisiRoutes(r *gin.Engine) {

	disposisi := r.Group("/disposisi")
	disposisi.Use(middlewares.AuthMiddleware())

	disposisi.POST("", middlewares.RoleMiddleware("Admin", "TU", "Kepala TU", "Kepala Sekolah"), controllers.CreateDisposisi)
	disposisi.POST("/approve", middlewares.RoleMiddleware("Kepala Sekolah"), controllers.ApproveDisposisi)
	disposisi.GET("/surat/:surat_id", controllers.GetDisposisi)
	disposisi.GET("", controllers.ListDisposisi)
}

func DashboardRoutes(r *gin.Engine) {
	dashboard := r.Group("/dashboard")
	dashboard.Use(middlewares.AuthMiddleware())

	dashboard.GET("", controllers.Dashboard)
}

func DocsRoutes(r *gin.Engine) {
	r.GET("/docs", func(c *gin.Context) {
		c.File("docs/index.html")
	})

	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("openapi.yaml")
	})
}