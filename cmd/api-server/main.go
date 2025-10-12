package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"mangahub/internal/config"
)

func main() {
	r := gin.Default()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	httpServer := fmt.Sprintf("127.0.0.1:%d", cfg.HTTPPort)

	r.Run(httpServer)
}
