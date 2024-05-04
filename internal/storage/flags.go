package storage

import (
	"flag"
	"fmt"
	"os"
)

var RunAddr string
var FlagRunAddr string
var defaultRunAddr = "localhost:8080"

var ReportInterval string
var FlagReportInterval string
var defaultReportInterval = "10"

var PollInterval string
var FlagPollInterval string
var defaultPollInterval = "2"

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&FlagReportInterval, "r", defaultReportInterval, "data report interval")
	flag.StringVar(&FlagPollInterval, "p", defaultPollInterval, "data poll interval")
	// парсим переданные серверу аргументы в зарегистрированные переменные

	flag.Parse()

	//os.Setenv("ADDRESS", "localhost:9999")
	RunAddr = setParam("ADDRESS", "a", defaultRunAddr)

	//os.Setenv("REPORT_INTERVAL", "333")
	ReportInterval = setParam("REPORT_INTERVAL", "r", defaultReportInterval)

	//os.Setenv("POLL_INTERVAL", "555")
	PollInterval = setParam("POLL_INTERVAL", "p", defaultPollInterval)
}

func setParam(envParamName, flagName, defaultValue string) string {

	retValue := ""
	if paramValue := os.Getenv(envParamName); paramValue != "" {
		//Если указана переменная окружения, то используется она.
		fmt.Println("Using environment value: " + paramValue)
		retValue = paramValue
	} else if IsFlagPassed(flagName) {
		//Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
		if flagName == "a" {
			fmt.Println("Using command line argument: " + FlagRunAddr)
			retValue = FlagRunAddr
		} else if flagName == "r" {
			fmt.Println("Using command line argument: " + FlagReportInterval)
			retValue = FlagReportInterval
		} else if flagName == "p" {
			fmt.Println("Using command line argument: " + FlagPollInterval)
			retValue = FlagPollInterval
		}

	} else {
		//Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.
		fmt.Println("Using default value: " + defaultValue)
		retValue = defaultValue
	}
	return retValue
}

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
