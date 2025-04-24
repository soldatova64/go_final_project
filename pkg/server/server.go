package server

import (
	"fmt"
	"github.com/soldatova64/go_final_project/pkg/api"
	"log"
	"net/http"
	"os"
	"strconv"
)

const port = 7540

func getPort() int {
	portStr := os.Getenv("TODO_PORT")
	if portStr == "" {
		return port
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(err)
	}
	return port
}

func Run() error {
	port := getPort()
	webDir := "./web"

	api.Init()

	fileServer := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileServer)

	log.Printf("Сервер запущен на порт: %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
