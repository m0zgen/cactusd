package main

import (
	"bufio"
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
	"regexp"
	"strings"
)

const MergedDir string = "merged"

var BufferSize int64

// Config file structure
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

// Config file loader
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

// Error handler
func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}

// If argument passed
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// If file exist in target
func isFileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	} else {
		return false
	}
}

// If two file is the same
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

// Detect runner form binary or form "go run"
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

// Auto config file path updater
func updatePath(filename string) string {
	var path string
	path = getWorkDir()
	filename = path + "/" + filename
	return filename
}

// Delete target file
func deleteFile(file string) {
	e := os.Remove(file)
	if e != nil {
		log.Fatal(e)
	}
}

// Thx: https://github.com/mactsouk/opensource.com
// Reference: https://opensource.com/article/18/6/copying-files-go
func copyFile(src, dst string, BUFFERSIZE int64) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file.", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if isFileExists(dst) {
		deleteFile(dst)
	}

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("File %s already exists.", dst)

	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}

// Create cataog in target place
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

// Get last octet in passed string
// Thx: https://github.com/peeyushsrj/golang-snippets
func getFilenameFromUrl(urlstr string) string {
	u, err := url.Parse(urlstr)
	if err != nil {
		log.Fatal("Error due to parsing url: ", err)
	}
	x, _ := url.QueryUnescape(u.EscapedPath())
	return filepath.Base(x)
}

// Merge files to one from folder to target
func mergeFiles(path string, ext string, dest string) {
	err := fileMerger.Merge(path, ext, dest, false)
	if err != nil {
		log.Fatal(err)
	}
}

// URL file downloader
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

	mergedFileName := MergedDir + "/" + getFilenameFromUrl(dest) + ".txt"
	if exist {
		matched, _ := isFileMatched(filepath, filepath+postfix)
		if matched {
			fmt.Println("Previous and current files - matched. No needed action.")
		} else {
			fmt.Printf("Merging files: %s\n", filename)

			if isFileExists(mergedFileName) {
				deleteFile(mergedFileName)
			}
			mergeFiles(dest, ".txt", mergedFileName)
		}
	} else {
		if isFileExists(mergedFileName) {
			deleteFile(mergedFileName)
		}
		mergeFiles(dest, ".txt", mergedFileName)
	}

	return nil
}

// URL iterator
func download(url []string, dest string) {
	//fmt.Println(url[1])
	for i, u := range url {
		fmt.Println(i, u)
		downloadFile(u, dest)
	}
}

func cleanFile(file string) {
	if isFileExists(file) {
		if err := os.Truncate(file, 0); err != nil {
			log.Printf("Failed to truncate: %v", err)
		}
	}
}

// Full regex - extract domain names (not used)
func fullRegex(file string, filename string, out string) {
	dat, err := os.Open(file)
	handleErr(err)
	defer dat.Close()

	scanner := bufio.NewScanner(dat)
	r1 := regexp.MustCompile(`(^#.*$)`)
	r2 := regexp.MustCompile(`(^(\/.*\/)$)|(^[a-z].*$)|(?:[\w-]+\.)+[\w-]+`)
	//re := regexp.MustCompile(`(?i)^(.*)(?:Inc\.|Incorp\.|Incorporation\.|Incorpa\.)(.*)$`)

	outFile := out + "/" + filename
	cleanFile(outFile)
	f, err := os.OpenFile(outFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	//
	for scanner.Scan() {

		if !r1.MatchString(scanner.Text()) {
			if _, err := f.WriteString(scanner.Text() + "\n"); err != nil {
				log.Println(err)
			}
			//fmt.Println(scanner.Text())
		}
		if r2.MatchString(scanner.Text()) {
			if _, err := f.WriteString(scanner.Text() + "\n"); err != nil {
				log.Println(err)
			}
			//fmt.Println(scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		handleErr(err)
	}

}

// Thx: https://gist.github.com/m0zgen/af44035db3102d08effc2d38e56c01f3
func prepareFiles(path string, fi os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if !!fi.IsDir() {
		return nil //
	}

	matched, err := filepath.Match("*.txt", fi.Name())

	if err != nil {
		panic(err)
		return err
	}

	replacer := strings.NewReplacer(
		"0.0.0.0 ", "",
		"0.0.0.0/8", "",
		"127.0.0.1", "",
		"=", "",
		"\n\n", "\n",
		" ", "",
	)

	r := regexp.MustCompile(`((?m)(^#|\s#).*)`)
	// Select empty lines
	r2 := regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	// Extract domain names from list:
	// (^(\/.*\/)$)|(^[a-z].*$)|(?:[\w-]+\.)+[\w-]+

	if matched {
		read, err := os.ReadFile(path)
		handleErr(err)
		//fmt.Println(string(read))
		fmt.Println(path)

		newContents := replacer.Replace(string(read))
		//newContents := strings.Replace(string(read), "0.0.0.0 ", "", -1)
		newContents = r.ReplaceAllString(newContents, "\n")
		newContents = r2.ReplaceAllString(newContents, "\n")
		//fmt.Println(newContents)

		err = os.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			panic(err)
		}

		fmt.Println(path)
	}

	return nil
}

// Process merged file
func publishFiles(mergeddir string, out string) {
	//// Process merged files
	files, err := os.ReadDir(MergedDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
		var plain bool = strings.Contains(file.Name(), "plain")
		var f string = mergeddir + "/" + file.Name()

		if plain {
			fmt.Println("Plain recurse for - " + f)
			//plainRegex(f, file.Name(), out)
		} else {
			fmt.Println("Full recurse for - " + file.Name())
			//fullRegex(f, file.Name(), out)
		}

		if err != nil {
			fmt.Printf("Invalid buffer size: %q\n", err)
			return
		}
		fmt.Println("Copy files from:" + f + " to: " + out + "/" + file.Name())
		err = copyFile(f, out+"/"+file.Name(), 20)
		if err != nil {
			fmt.Printf("File copying failed: %q\n", err)
		}
		fmt.Println("Publish files - Done!")
		//err := filepath.Walk(out, prepareFiles)
		//handleErr(err)
	}
}

// Main logic
func main() {

	// Get config and determine location
	var CONFIG string
	var dirStatus bool = strings.Contains(getWorkDir(), ".")

	// Get agrs
	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&CONFIG, "config", "config.yml", "Define config file")
	flag.Parse()
	if isFlagPassed("config") {
		fmt.Println(`Argument "-config" passed`)
	}

	// Load config
	config, _ := loadConfig(CONFIG, dirStatus)

	// Process catalogs & download
	//fmt.Println(config.Server.Port)
	createDir(MergedDir, dirStatus)
	createDir(config.Server.PublicDir, dirStatus)

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

	// Cleaning Process

	err := filepath.Walk(MergedDir, prepareFiles)
	handleErr(err)
	publishFiles(MergedDir, config.Server.PublicDir)
}
