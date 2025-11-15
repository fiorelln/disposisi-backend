package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/fiorelln/disposisi/controllers"
)

func AuthRoutes(r *gin.Engine) {
    auth := r.Group("/auth")
    {
        auth.POST("/register", controllers.Register)
        auth.POST("/login", controllers.Login)
    }
}
