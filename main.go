package main

import (
    "github.com/gin-gonic/gin"
    "github.com/fiorelln/disposisi/config"
    "github.com/fiorelln/disposisi/models"
    "github.com/fiorelln/disposisi/routes"

    "github.com/gin-contrib/cors" 
)

func main() {
    config.ConnectDB()
    config.DB.AutoMigrate(&models.User{})

    r := gin.Default()

      r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    }))

    routes.AuthRoutes(r)

    r.Run(":7000")
}
