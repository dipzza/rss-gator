package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user

	filepath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return err
	}

	return nil
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	filepath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	filepath = filepath + "/" + configFileName

	return filepath, nil
}

func Read() (Config, error) {
	var config Config

	filepath, err := getConfigFilePath()
	if err != nil {
		return config, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}