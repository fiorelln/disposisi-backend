package routes

import (
	"github.com/fiorelln/disposisi/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	AuthRoutes(r)
	AdminRoutes(r)
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

	admin.POST("/users", controllers.Register)
}