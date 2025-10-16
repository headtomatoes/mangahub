package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub/internal/config"
	"mangahub/database"

)

func main() {
	// r := gin.Default()
	// cfg, err := config.LoadConfig()
	// if err != nil {
	// 	log.Fatalf("could not load config: %v", err)
	// }
	// r.Use(gin.Logger())
	// r.Use(gin.Recovery())

	// httpServer := fmt.Sprintf("127.0.0.1:%d", cfg.HTTPPort)

	// r.GET("/check-conn", func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusOK, gin.H{
	// 		"message": "it's happening",
	// 	})
	// })

	// r.Run(httpServer)

	// 1️⃣ Connect to the database
	database.Connect()
	defer database.Close()

	// 2️⃣ Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// 3️⃣ Setup Gin
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/check-conn", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "API is alive and database connected ✅",
		})
	})

	httpServer := fmt.Sprintf("127.0.0.1:%d", cfg.HTTPPort)
	fmt.Println("🚀 Server running at", httpServer)
	r.Run(httpServer)
}
