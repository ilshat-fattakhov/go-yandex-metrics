package storage

import (
	"flag"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
var FlagRunAddr string

var FlagReportInterval string
var FlagPollInterval string

var ReportIntervalFlag uint64
var PollIntervalFlag uint64

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagReportInterval, "r", "10", "data report interval")
	flag.StringVar(&FlagPollInterval, "p", "2", "data poll interval")
	// парсим переданные серверу аргументы в зарегистрированные переменные

	flag.Parse()
}
