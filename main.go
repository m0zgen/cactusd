package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

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
	fmt.Println(filename)
	defer configFile.Close()
	if err != nil {
		return config, err
	}

	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	return config, err
}

func getCurrentDir() (string, error) {
	ex, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ex, err
	}
	path := filepath.Dir(ex)
	fmt.Println(path)
	return path, nil
}

//func checkDir(dir string) status, err {
//
//}

func downloadFile(file string, url string) {
	fmt.Println("Download file from link: " + url + " to: " + file)
}

func iterateUrls(url []string) {
	//fmt.Println(url[1])
	for i, u := range url {
		fmt.Println(i, u)
	}
}

func main() {
	var CONFIG string

	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&CONFIG, "config", "config.yml", "Enter path or config file name ")
	flag.Parse()

	config, _ := loadConfig(CONFIG)
	fmt.Println(config.Server.Port)
	//iterateUrls(config.Lists.Wl)
	getCurrentDir()
}
