package main

import (
	"go-yandex-metrics/cmd/agent/handlers"
	"log"
)

func main() {
	if err := handlers.RunAgent(); err != nil {
		log.Fatal(err)
	}
}
