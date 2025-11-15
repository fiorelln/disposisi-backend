package main

import (
    "github.com/gin-gonic/gin"
    "github.com/fiorelln/disposisi/config"
    "github.com/fiorelln/disposisi/models"
    "github.com/fiorelln/disposisi/routes"
)

func main() {
    config.ConnectDB()
    config.DB.AutoMigrate(&models.User{})

    r := gin.Default()

    routes.AuthRoutes(r)

    r.Run(":7000")
}
