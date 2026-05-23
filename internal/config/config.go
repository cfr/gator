package config

import (
    "encoding/json"
    "os"
    "io"
)

type Config struct {
    DbUrl string `json:"db_url"`
    CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
    home, err := os.UserHomeDir()
    if err == nil {
        return home + "/" + configFileName, nil
    } else {
        return "", err
    }
}

func write(config Config) error {
    bytes, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }
    path, err := getConfigFilePath()
    if err != nil {
        return nil
    }
    err = os.WriteFile(path, bytes, 0644)
    if err != nil {
        return err
    }
    return nil
}

func (config *Config) SetUser(user string) error {
    config.CurrentUserName = user
    err := write(*config)
    return err
}

func Read() (Config, error) {
    var config Config
    path, err := getConfigFilePath()
    if err != nil {
        return config, err
    }
    jsonFile, err := os.Open(path)
    if err != nil {
        return config, err
    }
    defer jsonFile.Close()

    bytes, err := io.ReadAll(jsonFile)
    if err != nil {
        return config, err
    }

    err = json.Unmarshal(bytes, &config)
    if err != nil {
        return config, err
    }

    return config, nil
}

