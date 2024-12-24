package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)


type Config struct {
	DbURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"


func (conf *Config) SetUser(username string) error{
	conf.CurrentUserName = username
	// Save to file
	if err := write(*conf); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}
// getConfigFilePath returns the full path to the config file in the user's home directory.
func getConfigFilePath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(wd, configFileName), nil
}

// write saves the given Config struct to the config file.
func write(cfg Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	return nil
}


func Read() (*Config, error){
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("faild to get home directory %w", err)
	}

	//Construct the full path to ~/.gatorconfig.json
	configPath := filepath.Join(wd, configFileName)

	//Open the JSON file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	//Decode the JSON file into a config 
	var conf Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&conf); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	return &conf, nil
}