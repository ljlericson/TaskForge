package main

import (
	"log"

	"github.com/ljlericson/TaskForge/internal/app"
)

func main() {
	taskForge, err := app.NewApp("config/server.yml")

	if err != nil {
		log.Fatal(err)
	}

	if err := taskForge.Run(); err != nil {
		log.Fatal(err)
	}
}
