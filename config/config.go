package config

import (
	"bufio"
	"fmt"
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

	NumProxies int `yaml:"numProxies"`

	TimeOut int `yaml:"timeout"`

	URLs []string

	Proxies []string
}

func GetConfig() (*Config, error) {
	yamlFile, err := os.ReadFile("config/example.yaml")
	if err != nil {
		return nil, fmt.Errorf("os read config file err: %s", err)
	}

	if err = yaml.Unmarshal(yamlFile, &C); err != nil {
		return nil, fmt.Errorf("unmarshl config file err: %s", err)
	}

	if C.Proxies, err = getProxies(); err != nil {
		return nil, fmt.Errorf("get proxies err: %s", err)
	}

	if C.URLs, err = getURLs(); err != nil {
		return nil, fmt.Errorf("get urls err: %s", err)
	}

	return &C, nil
}

func getURLs() ([]string, error) {
	file, err := os.Open("config/urls.txt")

	if err != nil {
		return nil, fmt.Errorf("open file err: %s", err)
	}

	defer func() {
		_ = file.Close()
	}()

	fileScanner := bufio.NewScanner(file)

	var urls []string
	for fileScanner.Scan() {
		urls = append(urls, fileScanner.Text())
	}

	if err := fileScanner.Err(); err != nil {
		return nil, fmt.Errorf("file scann err: %s", err)
	}

	return urls, nil
}

func getProxies() ([]string, error) {
	file, err := os.Open("config/proxies.txt")

	if err != nil {
		return nil, fmt.Errorf("open file err: %s", err)
	}

	defer func() {
		_ = file.Close()
	}()

	fileScanner := bufio.NewScanner(file)

	var proxies []string
	for fileScanner.Scan() {
		proxies = append(proxies, fileScanner.Text())
	}

	if err := fileScanner.Err(); err != nil {
		return nil, fmt.Errorf("file srann err: %s", err)
	}

	return proxies, nil
}
