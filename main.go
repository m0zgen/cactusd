package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Lists struct {
		Block []string `yaml:"block"`
	} `yaml:"lists"`
}

func loadConfig(filename string) (Config, error) {
	var config Config
	configFile, err := os.Open(filename)
	if err != nil {
		return config, err
	}

	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	return config, err
}

func main() {
	config, _ := loadConfig("test.yml")
	fmt.Println(config.Server.Port)
	fmt.Println(config.Lists.Block)
}
