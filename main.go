package main

import (
	"flag"
	"fmt"
	fileMerger "github.com/Ja7ad/goMerge"
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

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
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

func isFileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	} else {
		return false
	}
}

func isFileMatched(path1, path2 string) (sameSize bool, err error) {
	f1, err := os.Stat(path1)
	if err != nil {
		return
	}
	f2, err := os.Stat(path2)
	if err != nil {
		return
	}
	sameSize = f1.Size() == f2.Size()
	return
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

func mergeFiles(path string, ext string, dest string) {
	err := fileMerger.Merge(path, ext, dest, false)
	if err != nil {
		log.Fatal(err)
	}
}

func downloadFile(url string, dest string) error {
	var postfix string = "_prev"
	var filename string = getFilenameFromUrl(url)
	var filepath string = filepath.Join(dest, filename)
	if !strings.Contains(filename, ".txt") {
		filepath = filepath + ".txt"
	}

	// Check exists file for processing in future
	//if exists := getFileExists(filepath); exists == true {
	//	fmt.Printf("File exists %s\n", filename)
	//}

	exist := isFileExists(filepath)
	if exist {
		e := os.Rename(filepath, filepath+postfix)
		if e != nil {
			log.Fatal(e)
		}
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	fmt.Printf("Downloading file %s\n", filepath)
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

	if exist {
		matched, _ := isFileMatched(filepath, filepath+postfix)
		if matched {
			fmt.Println("Previous and current files - matched. No needed action.")
		} else {
			mergeFiles(dest, ".txt", "merged")
		}
	} else {
		mergeFiles(dest, ".txt", "merged/"+getFilenameFromUrl(dest)+".txt")
	}

	return nil
}

func download(url []string, dest string) {
	//fmt.Println(url[1])
	for i, u := range url {
		fmt.Println(i, u)
		downloadFile(u, dest)
	}
}

func main() {
	var CONFIG string
	var dirStatus bool = strings.Contains(getWorkDir(), ".")

	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&CONFIG, "config", "config.yml", "Define config file")
	flag.Parse()
	if isFlagPassed("config") {
		fmt.Println(`Argument "-config" passed`)
	}

	config, _ := loadConfig(CONFIG, dirStatus)

	//fmt.Println(config.Server.Port)
	createDir("merged", dirStatus)

	createDir(config.Server.DownloadDir+"/bl", dirStatus)
	download(config.Lists.Bl, config.Server.DownloadDir+"/bl")

	createDir(config.Server.DownloadDir+"/wl", dirStatus)
	download(config.Lists.Wl, config.Server.DownloadDir+"/wl")

	createDir(config.Server.DownloadDir+"/bl_plain", dirStatus)
	download(config.Lists.BlPlain, config.Server.DownloadDir+"/bl_plain")

	createDir(config.Server.DownloadDir+"/wl_plain", dirStatus)
	download(config.Lists.WlPlain, config.Server.DownloadDir+"/wl_plain")

	createDir(config.Server.DownloadDir+"/ip_plain", dirStatus)
	download(config.Lists.IpPlain, config.Server.DownloadDir+"/ip_plain")

	//

}
