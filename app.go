package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/routes"
	"github.com/joho/godotenv"
)

func main() {

	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		os.Exit(1)
	}

	// Register struct
	gob.Register(&models.UserInfo{})

	// Start DB
	if db.MigrateDb() != nil {
		log.Fatal("Failed to migrate database")
		os.Exit(1)
	}

	// Register Routes
	routes.Main()

	// Listen
	port := os.Getenv("SERVER_PORT")
	log.Println("Listening to localhost:" + port)
	http.ListenAndServe(fmt.Sprintf("localhost:%v", port), routes.Router)
}
