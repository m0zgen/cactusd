package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Server struct {
		Port           string `yaml:"port"`
		UpdateInterval int    `yaml:"update_interval"`
		DownloadDir    string `yaml:"download_dir"`
		UploadDir      string `yaml:"upload_dir"`
		PublicDir      string `yaml:"public_dir"`
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
		filename = updatePath(filename)
	}
	configFile, err := os.Open(filename)
	//fmt.Println(filename)
	defer configFile.Close()
	if err != nil {
		return config, err
	}

	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	return config, err
}

func getWorkDir() string {
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

func updatePath(filename string) string {
	var path string
	path = getWorkDir()
	filename = path + "/" + filename
	return filename
}

func createDir(dirName string, dirStatus bool) error {

	if !dirStatus {
		dirName = updatePath(dirName)
	}
	err := os.MkdirAll(dirName, os.ModeSticky|os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

// Thx: https://github.com/peeyushsrj/golang-snippets
func getFilenameFromUrl(urlstr string) string {
	u, err := url.Parse(urlstr)
	if err != nil {
		log.Fatal("Error due to parsing url: ", err)
	}
	x, _ := url.QueryUnescape(u.EscapedPath())
	return filepath.Base(x)
}

func downloadFile(url string) error {
	var filepath string = getFilenameFromUrl(url)
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func iterateUrls(url []string) {
	//fmt.Println(url[1])
	for i, u := range url {
		fmt.Println(i, u)
		downloadFile(u)
	}
}

func main() {
	var CONFIG string

	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&CONFIG, "config", "config.yml", "Enter path or config file name ")
	flag.Parse()

	dirStatus := strings.Contains(getWorkDir(), ".")
	config, _ := loadConfig(CONFIG, dirStatus)

	fmt.Println(config.Server.Port)

	createDir(config.Server.DownloadDir+"/bl", dirStatus)
	iterateUrls(config.Lists.Bl)

	//createDir(config.Server.DownloadDir+"/wl", dirStatus)
	//iterateUrls(config.Lists.Wl)
	//
	//createDir(config.Server.DownloadDir+"/bl_plain", dirStatus)
	//iterateUrls(config.Lists.BlPlain)
	//
	//createDir(config.Server.DownloadDir+"/wl_plain", dirStatus)
	//iterateUrls(config.Lists.WlPlain)
	//
	//createDir(config.Server.DownloadDir+"/ip_plain", dirStatus)
	//iterateUrls(config.Lists.IpPlain)

	//

}
