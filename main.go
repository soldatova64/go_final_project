package main

import (
	"github.com/joho/godotenv"
	"github.com/soldatova64/go_final_project/pkg/db"
	"github.com/soldatova64/go_final_project/pkg/server"
	"log"
	"os"
)

func getDBFile() string {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile != "" {
		return dbFile
	}
	return "scheduler.db"
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbfile := getDBFile()
	if err := db.Init(dbfile); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = server.Run()

	if err != nil {
		panic(err)
	}
}
