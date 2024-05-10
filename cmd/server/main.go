package main

import (
	"go-yandex-metrics/cmd/server/handlers"
	"log"
)

func main() {
	if err := handlers.RunServer(); err != nil {
		log.Fatal(err)
	}
}
