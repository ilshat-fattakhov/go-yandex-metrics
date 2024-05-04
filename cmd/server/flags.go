package main

import (
	"flag"
	"fmt"
	"go-yandex-metrics/internal/storage"
	"os"
)

var runAddr string
var flagRunAddr string
var defaultRunAddr = "localhost:8080"

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	//os.Setenv("ADDRESS", "localhost:9999")

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		//Если указана переменная окружения, то используется она.
		fmt.Println("Using environment value")
		runAddr = envRunAddr
	} else if storage.IsFlagPassed("a") {
		//Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
		fmt.Println("Using command line argument")
		runAddr = flagRunAddr

	} else {
		//Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.
		fmt.Println("Using default value")
		runAddr = defaultRunAddr

	}
}
