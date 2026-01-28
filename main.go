package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/config"
	"github.com/falasefemi2/companyflowlow/database"
)

func main() {
	fmt.Println("connecting to database")
	pool, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("Running migrations...")
	if err := database.RunMigrations(pool); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	router := mux.NewRouter()

	port := ":8080"
	fmt.Printf("\n✓ Server starting on http://localhost%s\n", port)
	fmt.Println("Press Ctrl+C to stop the server\n")

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
