package main

import (
	"log"

	"github.com/xdars/grpc-tasks/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
