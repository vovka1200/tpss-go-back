package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vovka1200/tpss-go-back/server"
	"os"
)

type Config struct {
	LogLevel string        `json:"loglevel"`
	Server   server.Server `json:"server"`
}

func (config *Config) read() {
	byteValue, err := os.ReadFile("config.json")
	if err == nil {
		err = json.Unmarshal(byteValue, &config)
		if err == nil {
			log.WithFields(log.Fields{
				"loglevel": config.LogLevel,
				"listen":   config.Server.Listen,
			}).Info("Config")
		} else {
			log.Fatal(fmt.Sprintf("Ошибка чтения конфигурации: %s", err.Error()))
		}
	} else {
		log.Fatal(fmt.Sprintf("Ошибка чтения файла конфигурации: %s", err.Error()))
	}
}
