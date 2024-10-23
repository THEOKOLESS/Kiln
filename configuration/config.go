package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

type MainConfig struct {
	DatabaseName string `json:"database"`
	Limit        string `json:"limit"`
	AllData      bool   `json:"all_data"`
}

func New(configPath string) (MainConfig, error) {
	ps, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Config ReadFile error:", err)
		return MainConfig{}, err
	}
	var configuration MainConfig
	err = json.Unmarshal(ps, &configuration)
	if err != nil {
		fmt.Println("Config Unmarshal error:", err)
		return MainConfig{}, err
	}
	return configuration, nil
}

func Init(configPath string) (MainConfig, error) {
	if configPath == "" {
		fmt.Println("Please provide the configuration file")
		os.Exit(100)
	}
	config, err := New(configPath)
	if err != nil {
		return MainConfig{}, err
	}
	return config, nil
}
