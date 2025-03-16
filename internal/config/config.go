package config

import (
	"os"
	"io"
	"encoding/json"
)

// I do not understand why it only works if I have the converter do it
// but whatever it's... it's fine...
type Config struct {
	DbURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

// get the file path to the config
// this feels too hard-coded but I'll figure that out later
const configFileName = ".gatorconfig.json"
const configFilePath = "/workspace/github.com/katheland/blogaggregator/"
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir + configFilePath + configFileName, nil
}

// read the config json and make a Config struct from it
func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	
	jsonData, err := io.ReadAll(file)
	if err != nil {
		return Config{}, err
	}
	
	var config Config
	if err = json.Unmarshal(jsonData, &config); err != nil {
		return Config{}, err
	}
	
	return config, nil
}

// set the current username
func (c* Config) SetUser(username string) {
	c.CurrentUserName = username
	write(c)
}

// overwrites the config file
func write(cfg* Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(string(jsonData))
	return nil
}