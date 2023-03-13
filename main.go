package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
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

func loadConfig(filename string, dirStatus bool) (Config, error) {
	var config Config
	// Check go run or run binary
	if !dirStatus {
		var path string
		path = GetWorkDir()
		filename = path + "/" + filename
	}
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

func checkDir(dirName string) error {

	err := os.MkdirAll(dirName, os.ModeDir)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

func downloadFile(file string, url string) {
	fmt.Println("Download file from link: " + url + " to: " + file)
}

func iterateUrls(url []string) {
	//fmt.Println(url[1])
	for i, u := range url {
		fmt.Println(i, u)
	}
}

func GetWorkDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	dir := filepath.Dir(ex)

	// Helpful when developing:
	// when running `go run`, the executable is in a temporary directory.
	if strings.Contains(dir, "go-build") {
		return "."
	}
	return filepath.Dir(ex)
}

func main() {
	var CONFIG string

	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&CONFIG, "config", "config.yml", "Enter path or config file name ")
	flag.Parse()

	dirStatus := strings.Contains(GetWorkDir(), ".")
	config, _ := loadConfig(CONFIG, dirStatus)

	fmt.Println(config.Server.Port)
	//checkDir("download-aaaaa")
	//iterateUrls(config.Lists.Wl)
	//getCurrentDir()
	fmt.Println(GetWorkDir())
}
