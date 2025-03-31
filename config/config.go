package config

import (
	"encoding/json"
	"os"
	"sync"
)

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	SMTPConfig     `json:"smtp"`
}

var (
	mu         sync.Mutex
	smtpConfig SMTPConfig
)

func GetSMTPConfig() SMTPConfig {
	mu.Lock()
	defer mu.Unlock()
	return smtpConfig
}

func loadConfig() error {
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var appConfig Config
	err = decoder.Decode(&appConfig)
	if err != nil {
		return err
	}
	smtpConfig = appConfig.SMTPConfig
	return nil
}

func InitConfig() error {
	err := loadConfig()
	return err
}