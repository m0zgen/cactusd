package main

import (
	"bufio"
	"flag"
	"fmt"
	fileMerger "github.com/Ja7ad/goMerge"
	"gopkg.in/yaml.v3"
	"html/template"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var CONFIG string
var hostsPingStat = make(map[string]string)

const MergedDir string = "merged"

var BufferSize int64

// Config file structure
type Config struct {
	Server struct {
		Port           string `yaml:"port"`
		UpdateInterval string `yaml:"update_interval"`
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

// Cfg Testing in future
type Cfg struct {
	serverConfig map[string]Server
	listsConfig  map[string]interface{}
	pingConfig   map[string]interface{}
}

// Server Testing in future
type Server struct {
	Port           string
	UpdateInterval string
	DownloadDir    string
	UploadDir      string
	PublicDir      string
}

type PingHost struct {
	name string
	port int
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

func loadUnmarshalConfig(filename string, dirStatus bool) map[string]interface{} {

	// Check go run or run binary
	if !dirStatus {
		filename = updatePath(filename)
	}
	configFile, err := os.ReadFile(filename)
	//fmt.Println(filename)
	handleErr(err)

	var data map[string]interface{}
	err = yaml.Unmarshal(configFile, &data)
	handleErr(err)

	return data
}

// func getDatetime() time.Time
func getTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Error handler
func handleErr(e error) {
	if e != nil {
		//panic(e)
		log.Println(e)
	}
}

// Sigterm handler
func handler(signal os.Signal) {
	if signal == syscall.SIGTERM {
		fmt.Println("Got kill signal. ")
		fmt.Println("Program will terminate now.")
		os.Exit(0)
	} else if signal == syscall.SIGINT {
		fmt.Println("Got CTRL+C signal")
		fmt.Println("Closing.")
		os.Exit(0)
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

func isDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true
	}
	return false
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

// Create catalog in target place
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
	var postfix = "_prev"
	var filename = getFilenameFromUrl(url)
	var filepath = filepath.Join(dest, filename)
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
	// TODO: detect 404 pages or 200 response
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

// Thx: https://stackoverflow.com/questions/7424340/read-in-lines-in-a-text-file-sort-then-overwrite-file/7425283#7425283
func readLines(file string) (lines []string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		const delim = '\n'
		line, err := r.ReadString(delim)
		if err == nil || len(line) > 0 {
			if err != nil {
				line += string(delim)
			}
			lines = append(lines, line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return lines, nil
}

func writeLines(file string, lines []string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	for _, line := range lines {
		_, err := w.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil
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
		"127.0.0.1 ", "",
		"\n\n", "\n",
		"=", "",
		" ", "",
	)

	//r := regexp.MustCompile(`((?m)(#|\s#).*)`)
	r2 := regexp.MustCompile(`(?m)(^[$&+,:;=?@#|'<>.\-^*()%!\s\d{0,9}].+$)|(^.*::.*)|(#|\s#.*)|(^.*\/\/.*)`)
	// Select empty lines
	//r2 := regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	// Extract domain names from list:
	// (^(\/.*\/)$)|(^[a-z].*$)|(?:[\w-]+\.)+[\w-]+

	if matched {
		read, err := os.ReadFile(path)
		handleErr(err)
		//fmt.Println(string(read))
		fmt.Println(path)

		newContents := replacer.Replace(string(read))
		//newContents := strings.Replace(string(read), "0.0.0.0 ", "", -1)
		//newContents = r.ReplaceAllString(newContents, "\n")
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

// Sort and remove duplicates from files
func sortFile(file string) {

	lines, err := readLines(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	removeMatches(&lines)
	RemoveDuplicates(&lines)
	sort.Strings(lines)
	err = writeLines(file, lines)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func removeMatches(lines *[]string) {
	//r := regexp.MustCompile(`^\s*$[\r\n]*|[\r\n]+\s+\z`)
	r := regexp.MustCompile(`((#|\s#).*)|(^\s*$[\r\n]*|[\r\n]+\s+\z)|(^\d{0,9}$)`)
	j := 0
	for index := range *lines {

		(*lines)[j] = r.ReplaceAllString((*lines)[index], "\n")
		j++
	}
	*lines = (*lines)[:j]
}

func RemoveDuplicates(lines *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *lines {
		if !found[x] {
			found[x] = true
			(*lines)[j] = (*lines)[i]
			j++
		}
	}
	*lines = (*lines)[:j]
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

		var plain = strings.Contains(file.Name(), "plain")
		var f = mergeddir + "/" + file.Name()

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

		sortFile(f)

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

func initial(config Config, dirStatus bool) {

	// Folder name for published file, public sub-catalog
	var publishFilesDir = config.Server.PublicDir + "/files"
	// Process catalogs & download
	//fmt.Println(config.Server.Port)
	createDir(MergedDir, dirStatus)
	createDir(config.Server.UploadDir, dirStatus)
	createDir(config.Server.PublicDir, dirStatus)
	createDir(publishFilesDir, dirStatus)

	// Download files
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
	publishFiles(MergedDir, publishFilesDir)
	//
	if !isDirEmpty(config.Server.UploadDir) {
		err := filepath.Walk(config.Server.UploadDir, prepareFiles)
		handleErr(err)

		mergedFileName := publishFilesDir + "/dropped_ip.txt"
		mergeFiles(config.Server.UploadDir, ".txt", mergedFileName)
		sortFile(mergedFileName)
	}

}

func runTicker(config Config, dirStatus bool, group *sync.WaitGroup) {
	defer group.Done()

	initial(config, dirStatus)
	callPinger()
	fmt.Println("Interval done at: " + getTime())
	fmt.Println("Next iteration will start after: " + config.Server.UpdateInterval)

	duration, err := time.ParseDuration(config.Server.UpdateInterval)
	handleErr(err)
	tick := time.Tick(time.Duration(duration.Minutes()) * time.Minute)

	for range tick {
		initial(config, dirStatus)
		callPinger()
		fmt.Println("Interval done at: " + getTime())
		fmt.Println("Next iteration will start after: " + config.Server.UpdateInterval)
	}
}

func webUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file") // the FormFile function takes in the POST input id file

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	//

	f, err := os.OpenFile("./upload/"+header.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	handleErr(err)

	defer f.Close()
	io.Copy(f, file)

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)

}

func runHttpServer(port string) {

	//fileHandler := http.StripPrefix("/", http.FileServer(http.Dir("public")))

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", serveTemplate)

	//http.Handle("/", fileHandler)
	http.HandleFunc("/time", timeHandler)
	http.HandleFunc("/upload", webUploadHandler)

	// Run server
	//handler := http.FileServer(http.Dir("./public"))
	//http.Handle("/download", handler)

	log.Print("Listening on : " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Thx: https://stackoverflow.com/questions/1094841/get-human-readable-version-of-file-size/1094933#1094933
func prettyByteSize(b int) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

// type PublicFiles
func listPublicFilesDir(target string) map[string]string {
	files, err := os.ReadDir(target)
	m := make(map[string]string)
	var sz string
	var PublicFiles []string
	if err != nil {
		handleErr(err)
	}

	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())
		//fmt.Println(target + file.Name())
		i, err := file.Info()
		handleErr(err)
		sz = prettyByteSize(int(i.Size()))
		m[file.Name()] = sz
		PublicFiles = append(PublicFiles, file.Name(), sz)
	}

	//return PublicFiles
	return m
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {

	appVersion := "0.1.7"
	hostname, err := os.Hostname()
	handleErr(err)
	publicFiles := listPublicFilesDir("./public/files/")

	//
	files := []string{
		"./templates/base.html",
		"./templates/partials/nav.html",
		"./templates/home.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	////
	data := struct {
		AppVersion  string
		CurrentDate string
		HostName    string
		PublicFiles map[string]string
		PingStatus  map[string]string
	}{
		appVersion,
		getTime(),
		hostname,
		publicFiles,
		hostsPingStat,
	}

	//err = ts.Execute(w, data)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//}
	////

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}

}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, time.Now().Format("02 Jan 2006 15:04:05 MST"))
}

func pingHost(host string, p int) {
	port := strconv.Itoa(p)
	timeout := time.Duration(2 * time.Second)
	_, err := net.DialTimeout("tcp", host+":"+port, timeout)
	if err != nil {
		fmt.Printf("%s %s %s\n", host, "not responding", err.Error())
		hostsPingStat[host+" ("+port+")"] = "Not response"
		//return host + " (" + port + ")", false
	} else {
		fmt.Printf("%s %s %s\n", host, "responding on port:", port)
		hostsPingStat[host+" ("+port+")"] = "Ok"
		//return host + " (" + port + ")", true
	}

}

// Main logic
func main() {

	// Get config and determine location

	var dirStatus = strings.Contains(getWorkDir(), ".")
	var wg = new(sync.WaitGroup)
	wg.Add(4)

	// Get arguments
	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&CONFIG, "config", "config.yml", "Define config file")
	flag.Parse()
	if isFlagPassed("config") {
		fmt.Println(`Argument "-config" passed`)
	}

	config, _ := loadConfig(CONFIG, dirStatus)

	//serverConfig := configData["server"].(map[string]interface{})
	//listsConfig := configData["lists"].(map[string]interface{})

	//log.Println(configData)
	//for k, v := range configData {
	//	log.Println(k, ":", v)
	//
	//}

	//fmt.Println(reflect.TypeOf(config))

	// Routines start
	// Will test multiple servers
	//go func() {
	//	server := createServer(3301, "Server 1")
	//	fmt.Println(server.ListenAndServe())
	//	wg.Done()
	//}()
	//
	//go func() {
	//	server := createServer(3302, "Server 2")
	//	fmt.Println(server.ListenAndServe())x
	//	wg.Done()
	//}()

	go runHttpServer(config.Server.Port)
	go runTicker(config, dirStatus, wg)

	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)
	exitchnl := make(chan int)

	// Handle interrupt signals
	// Thx: https://www.developer.com/languages/os-signals-go/
	go func() {
		for {
			s := <-sigchnl
			handler(s)
		}
	}()

	exitcode := <-exitchnl
	os.Exit(exitcode)

	wg.Wait()
	// Routines end

}

// Testing
func callPinger() {

	var dirStatus = strings.Contains(getWorkDir(), ".")

	configData := loadUnmarshalConfig(CONFIG, dirStatus)
	pingConfig := configData["ping"].([]interface{})
	var p PingHost
	for _, v := range pingConfig {
		//log.Println(k, ":", v)
		targets := v.(map[string]interface{})
		for _, param := range targets {
			//fmt.Println(param)
			hosts := param.(map[string]interface{})
			for options, host := range hosts {
				switch options {
				case "name":
					p.name = host.(string)
				case "port":
					p.port = host.(int)
				}

			}

		}
		//fmt.Println("Target: ", p.name)
		pingHost(p.name, p.port)
	}

}
