package main

import (
    "github.com/gin-gonic/gin"
    "github.com/fiorelln/disposisi/config"
)

func main() {
    config.ConnectDatabase()

    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Backend Disposisi berjalan!",
        })
    })

    r.Run(":8080")
}
