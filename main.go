package main

import (
	"fmt"
	"os"

	_ "time/tzdata"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Initialize the application (loads config from APP_ENV)
	app := InitializeApp()

	// Start the application and server
	message := app.Start()
	fmt.Println(message)

	// Wait for shutdown signal (SIGINT or SIGTERM)
	fmt.Println("Application running. Press Ctrl+C to shutdown...")
	app.Wait()

	// Graceful shutdown
	fmt.Println("Initiating graceful shutdown...")
	if err := app.Stop(); err != nil {
		fmt.Printf("Error during application stop: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application shutdown complete")
}
