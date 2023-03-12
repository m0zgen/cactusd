package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const CONFIG string = "config.yml"

type Config struct {
	Server struct {
		Port           string `yaml:"port"`
		UpdateInterval int    `yaml:"updateInterval"`
		DownloadDir    string `yaml:"downloadDir"`
		UploadDir      string `yaml:"upploadDir"`
		PublicDir      string `yaml:"publicDir"`
	} `yaml:"server"`
	Lists struct {
		Bl      []string `yaml:"bl"`
		BlPlain []string `yaml:"bl_plain"`
		Wl      []string `yaml:"wl"`
		WlPlain []string `yaml:"wl_plain"`
		IpPlain []string `yaml:"ip_plain"`
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
	config, _ := loadConfig(CONFIG)
	fmt.Println(config.Server.Port)
	fmt.Println(config.Lists.Bl)
}
