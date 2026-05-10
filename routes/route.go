package routes

import (
	"github.com/fiorelln/disposisi/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {

	auth := r.Group("/auth")
	{
		auth.POST("/login", controllers.Login)

		auth.POST("/forgot-password", controllers.ForgotPassword)

		auth.POST("/verify-otp", controllers.VerifyOTP)

		auth.POST("/reset-password", controllers.ResetPassword)
	}

	admin := r.Group("/admin")
	{
		admin.POST("/users", controllers.Register)
	}
}