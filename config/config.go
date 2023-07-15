package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

var C Config

type Config struct {
	DB struct {
		Host string `yaml:"host"` // Хост
		Port string `yaml:"port"` // Порт
		User string `yaml:"user"` // Пользователь
		Pass string `yaml:"pass"` // Пароль
		Name string `yaml:"name"` // Название
	} `yaml:"db"`

	URLs []string

	Proxies []string
}

const configFile = "config/example.yaml"

func GetConfig() (*Config, error) {
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.New("read config fIle err: " + err.Error())
	}

	if err = yaml.Unmarshal(yamlFile, &C); err != nil {
		return nil, errors.New("unmarshal config fIle err: " + err.Error())
	}

	return &C, nil
}

func getURLs() {

}

func getProxies() {

}
