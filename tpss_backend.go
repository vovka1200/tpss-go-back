package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

var app struct {
	Config
}

func main() {
	var err error
	var logLevel log.Level
	// Чтение конфигурации
	log.SetOutput(os.Stdout)
	app.Config.read()
	if logLevel, err = log.ParseLevel(app.Config.LogLevel); err != nil {
		log.Fatal(fmt.Sprintf("Ошибка журналирования: %s(пропущено значение)", err.Error()))
	}
	// Настройка журналирования
	log.SetLevel(logLevel)
	if logLevel == log.DebugLevel || logLevel == log.TraceLevel {
		log.SetReportCaller(true)
	}

	app.Config.Server.Run()

}
