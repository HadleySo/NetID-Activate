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
	"github.com/spf13/viper"
)

func main() {

	// LOAD CONFIG
	viper.SetConfigName("netid")

	// Config search paths to find the file
	viper.AddConfigPath(".")
	viper.AddConfigPath("./data")

	err := viper.ReadInConfig()

	// Viper errors
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// Viper prefix for environment variables
	viper.SetEnvPrefix("NETID")
	viper.AutomaticEnv()

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
	port := viper.GetString("SERVER_PORT")
	log.Println("Listening to localhost:" + port)
	http.ListenAndServe(fmt.Sprintf("localhost:%v", port), routes.Router)
}
