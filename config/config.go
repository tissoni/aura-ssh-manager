package config

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	Theme string `json:"theme"`
}

var DefaultConfig = Config{
	Theme: "aura",
}

func GetConfigPath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, "Aura", "config.json")
}

func Load() Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig
	}

	var conf Config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return DefaultConfig
	}
	return conf
}

func Save(conf Config) error {
	path := GetConfigPath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
