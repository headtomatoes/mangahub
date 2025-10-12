package main

import (
	"fmt"
	"log"

	"mangahub/internal/config"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	// Use config throughout your application
	fmt.Println(config)

}
