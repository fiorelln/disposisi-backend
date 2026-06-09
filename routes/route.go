package routes

import (
	"github.com/fiorelln/disposisi/controllers"
	"github.com/fiorelln/disposisi/middlewares"
	"github.com/gin-gonic/gin"
)

var (
	SuratMasukCtrl  *controllers.SuratMasukController
	SuratKeluarCtrl *controllers.SuratKeluarController
	DisposisiCtrl   *controllers.DisposisiController
)

func SetupRoutes(r *gin.Engine) {
	AuthRoutes(r)
	SuratMasukRoutes(r)
	SuratKeluarRoutes(r)
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

func SuratMasukRoutes(r *gin.Engine) {
	sm := r.Group("/surat-masuk")
	sm.Use(middlewares.AuthMiddleware())

	sm.POST("", middlewares.RoleMiddleware("admin", "Tata Usaha"), SuratMasukCtrl.Register)
	sm.POST("/:id/forward-to-principal", middlewares.RoleMiddleware("admin", "Tata Usaha"), SuratMasukCtrl.ForwardToPrincipal)
	sm.POST("/:id/review", middlewares.RoleMiddleware("kepala sekolah"), SuratMasukCtrl.Review)
	sm.POST("/:id/distribute", middlewares.RoleMiddleware("admin", "Tata Usaha"), SuratMasukCtrl.DistributeToUser)
	sm.GET("/:id", SuratMasukCtrl.GetByID)
	sm.GET("", SuratMasukCtrl.List)
	sm.GET("/inbox/distribusi", SuratMasukCtrl.GetInboxDistribusi)
}

func SuratKeluarRoutes(r *gin.Engine) {
	sk := r.Group("/surat-keluar")
	sk.Use(middlewares.AuthMiddleware())

	sk.POST("", middlewares.RoleMiddleware("admin", "Tata Usaha"), SuratKeluarCtrl.Create)
	sk.POST("/:id/submit-to-principal", middlewares.RoleMiddleware("admin", "Tata Usaha"), SuratKeluarCtrl.SubmitToPrincipal)
	sk.POST("/:id/review", middlewares.RoleMiddleware("kepala sekolah"), SuratKeluarCtrl.Review)
	sk.GET("/:id", SuratKeluarCtrl.GetByID)
	sk.GET("", SuratKeluarCtrl.List)
}

func DisposisiRoutes(r *gin.Engine) {
	d := r.Group("/disposisi")
	d.Use(middlewares.AuthMiddleware())

	d.GET("/inbox", DisposisiCtrl.GetInbox)
	d.GET("/sent", DisposisiCtrl.GetSentItems)
	d.GET("/stats", DisposisiCtrl.GetStats)
	d.GET("/:id", DisposisiCtrl.GetDetail)
	d.POST("/:id/read", DisposisiCtrl.MarkAsRead)
	d.POST("/:id/waka-action", middlewares.RoleMiddleware("waka kesiswaan", "waka kurikulum", "waka sarpras", "waka humas", "bk", "kapro rpl", "kapro tkj", "kapro dkv", "kapro an", "kapro ei", "kapro mt", "kapro av", "kapro bc", "bkk", "prakerin"), DisposisiCtrl.WakaAction)
	d.POST("/:id/complete", DisposisiCtrl.CompleteDisposisi)

	// History of a surat
	sm := r.Group("/surat")
	sm.Use(middlewares.AuthMiddleware())
	sm.GET("/:surat_id/history", DisposisiCtrl.GetHistory)
}

func DashboardRoutes(r *gin.Engine) {
	d := r.Group("/dashboard")
	d.Use(middlewares.AuthMiddleware())
	d.GET("", controllers.Dashboard)
}

func DocsRoutes(r *gin.Engine) {
	r.GET("/docs", func(c *gin.Context) {
		c.File("docs/index.html")
	})
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("openapi.yaml")
	})
}
