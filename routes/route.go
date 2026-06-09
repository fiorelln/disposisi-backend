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
	auth := r.Group("/auth")
	auth.POST("/login", controllers.Login)
	auth.POST("/forgot-password", controllers.ForgotPassword)
	auth.POST("/verify-otp", controllers.VerifyOTP)
	auth.POST("/reset-password", controllers.ResetPassword)

	registerPlatformRoutes(r)
}

func registerPlatformRoutes(r *gin.Engine) {
	platforms := []string{
		"surat-masuk",
		"surat-keluar",
		"disposisi",
		"dashboard",
	}

	for _, p := range platforms {
		registerPlatform(r, "api/v1/desktop/"+p)
		registerPlatform(r, "api/v1/web/"+p)
		registerPlatform(r, "api/v1/mobile/"+p)
	}

	registerHistory(r.Group("/surat"))
	registerDocs(r)
}

func registerPlatform(r *gin.Engine, base string) {
	g := r.Group("/" + base)
	g.Use(middlewares.AuthMiddleware())

	switch {
	case contains(base, "surat-masuk"):
		registerSuratMasuk(g)
	case contains(base, "surat-keluar"):
		registerSuratKeluar(g)
	case contains(base, "disposisi"):
		registerDisposisi(g)
	case contains(base, "dashboard"):
		g.GET("", controllers.Dashboard)
	}
}

func registerSuratMasuk(g *gin.RouterGroup) {
	tu := []string{"admin", "Tata Usaha"}
	g.POST("", middlewares.RoleMiddleware(tu...), SuratMasukCtrl.Register)
	g.POST("/:id/forward-to-principal", middlewares.RoleMiddleware(tu...), SuratMasukCtrl.ForwardToPrincipal)
	g.POST("/:id/review", middlewares.RoleMiddleware("kepala sekolah"), SuratMasukCtrl.Review)
	g.POST("/:id/distribute", middlewares.RoleMiddleware(tu...), SuratMasukCtrl.DistributeToUser)
	g.GET("/:id", SuratMasukCtrl.GetByID)
	g.GET("/:id/download", SuratMasukCtrl.Download)
	g.GET("", SuratMasukCtrl.List)
	g.GET("/inbox/distribusi", SuratMasukCtrl.GetInboxDistribusi)
}

func registerSuratKeluar(g *gin.RouterGroup) {
	tu := []string{"admin", "Tata Usaha"}
	g.POST("", middlewares.RoleMiddleware(tu...), SuratKeluarCtrl.Create)
	g.POST("/:id/submit-to-principal", middlewares.RoleMiddleware(tu...), SuratKeluarCtrl.SubmitToPrincipal)
	g.POST("/:id/review", middlewares.RoleMiddleware("kepala sekolah"), SuratKeluarCtrl.Review)
	g.POST("/:id/finalize", middlewares.RoleMiddleware(tu...), SuratKeluarCtrl.Finalize)
	g.GET("/:id", SuratKeluarCtrl.GetByID)
	g.GET("/:id/download", SuratKeluarCtrl.Download)
	g.GET("", SuratKeluarCtrl.List)
}

func registerDisposisi(g *gin.RouterGroup) {
	wakaRoles := []string{"waka kesiswaan", "waka kurikulum", "waka sarpras", "waka humas", "bk",
		"kapro rpl", "kapro tkj", "kapro dkv", "kapro an", "kapro ei",
		"kapro mt", "kapro av", "kapro bc", "bkk", "prakerin"}
	g.GET("/inbox", DisposisiCtrl.GetInbox)
	g.GET("/sent", DisposisiCtrl.GetSentItems)
	g.GET("/stats", DisposisiCtrl.GetStats)
	g.GET("/:id", DisposisiCtrl.GetDetail)
	g.POST("/:id/read", DisposisiCtrl.MarkAsRead)
	g.POST("/:id/waka-action", middlewares.RoleMiddleware(wakaRoles...), DisposisiCtrl.WakaAction)
	g.POST("/:id/complete", DisposisiCtrl.CompleteDisposisi)
}

func registerHistory(g *gin.RouterGroup) {
	g.Use(middlewares.AuthMiddleware())
	g.GET("/:surat_id/history", DisposisiCtrl.GetHistory)
}

func registerDocs(r *gin.Engine) {
	r.GET("/docs", func(c *gin.Context) {
		c.File("docs/index.html")
	})
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("openapi.yaml")
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
