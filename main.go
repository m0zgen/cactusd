package main

import (
	"bufio"
	"cactusd/util"
	conf "cactusd/util"
	"flag"
	"fmt"
	fileMerger "github.com/Ja7ad/goMerge"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var HostsPingStat = conf.HostsPingStat

const MergedDir = conf.MergedDir

var BufferSize int64

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

func responseOutput(w http.ResponseWriter, message string) (int, error) {
	return fmt.Fprint(w, message)
}

// Thx: https://github.com/mactsouk/opensource.com
// Reference: https://opensource.com/article/18/6/copying-files-go
func copyFile(src, dst string, BUFFERSIZE int64) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if util.IsFileExists(dst) {
		util.DeleteFile(dst)
	}

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("file %s already exists", dst)

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
	var filename = util.GetFilenameFromUrl(url)
	var filepath = filepath.Join(dest, filename)
	if !strings.Contains(filename, ".txt") {
		filepath = filepath + ".txt"
	}

	// Check exists file for processing in future
	//if exists := getFileExists(filepath); exists == true {
	//	fmt.Printf("File exists %s\n", filename)
	//}

	exist := util.IsFileExists(filepath)
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

	mergedFileName := MergedDir + "/" + util.GetFilenameFromUrl(dest) + ".txt"
	if exist {
		matched, _ := util.IsFileMatched(filepath, filepath+postfix)
		if matched {
			fmt.Println("Previous and current files - matched. No needed action.")
		} else {
			fmt.Printf("Merging files: %s\n", filename)

			if util.IsFileExists(mergedFileName) {
				util.DeleteFile(mergedFileName)
			}
			mergeFiles(dest, ".txt", mergedFileName)
		}
	} else {
		if util.IsFileExists(mergedFileName) {
			util.DeleteFile(mergedFileName)
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
		err := downloadFile(u, dest)
		util.HandleErr(err)
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
	r2 := regexp.MustCompile(
		`(?m)(^[$&+,:;=?@#|'<>.\-^*()%!].+$)|(^.*::.*)|(#|\s#.*)|(^.*\/\/.*)|(^.*,.*$)|(^.*\.-.*$)|(^.*[\$\^].*$)`)
	// Select empty lines
	//r2 := regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	// Extract domain names from list:
	// (^(\/.*\/)$)|(^[a-z].*$)|(?:[\w-]+\.)+[\w-]+

	if matched {
		read, err := os.ReadFile(path)
		util.HandleErr(err)
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

func initial(config conf.Config, dirStatus bool) {

	// Folder name for published file, public sub-catalog
	var publishFilesDir = config.Server.PublicDir + "/files"
	var err error
	// Process catalogs & download
	//fmt.Println(config.Server.Port)
	err = util.CreateDir(MergedDir, dirStatus)
	err = util.CreateDir(config.Server.UploadDir, dirStatus)
	err = util.CreateDir(config.Server.PublicDir, dirStatus)
	err = util.CreateDir(publishFilesDir, dirStatus)

	// Download files
	err = util.CreateDir(config.Server.DownloadDir+"/bl", dirStatus)
	download(config.Lists.Bl, config.Server.DownloadDir+"/bl")

	err = util.CreateDir(config.Server.DownloadDir+"/wl", dirStatus)
	download(config.Lists.Wl, config.Server.DownloadDir+"/wl")

	err = util.CreateDir(config.Server.DownloadDir+"/bl_plain", dirStatus)
	download(config.Lists.BlPlain, config.Server.DownloadDir+"/bl_plain")

	err = util.CreateDir(config.Server.DownloadDir+"/wl_plain", dirStatus)
	download(config.Lists.WlPlain, config.Server.DownloadDir+"/wl_plain")

	err = util.CreateDir(config.Server.DownloadDir+"/ip_plain", dirStatus)
	download(config.Lists.IpPlain, config.Server.DownloadDir+"/ip_plain")

	// Cleaning Process
	err = filepath.Walk(MergedDir, prepareFiles)
	util.HandleErr(err)
	publishFiles(MergedDir, publishFilesDir)
	//
	if !util.IsDirEmpty(config.Server.UploadDir) {
		err := filepath.Walk(config.Server.UploadDir, prepareFiles)
		util.HandleErr(err)

		mergedFileName := publishFilesDir + "/dropped_ip.txt"
		mergeFiles(config.Server.UploadDir, ".txt", mergedFileName)
		sortFile(mergedFileName)
	}

}

func runTicker(config conf.Config, dirStatus bool, group *sync.WaitGroup) {
	defer group.Done()

	initial(config, dirStatus)
	util.CallPinger()
	fmt.Println("Interval done at: " + util.GetTime())
	fmt.Println("Next iteration will start after: " + config.Server.UpdateInterval)

	duration, err := time.ParseDuration(config.Server.UpdateInterval)
	util.HandleErr(err)
	tick := time.Tick(time.Duration(duration.Minutes()) * time.Minute)

	for range tick {
		initial(config, dirStatus)
		util.CallPinger()
		fmt.Println("Interval done at: " + util.GetTime())
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
	util.HandleErr(err)

	defer f.Close()
	_, err = io.Copy(f, file)

	_, err = responseOutput(w, "File uploaded successfully : ")
	_, err = responseOutput(w, header.Filename)
	util.HandleErr(err)

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
		util.HandleErr(err)
	}

	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())
		//fmt.Println(target + file.Name())
		i, err := file.Info()
		util.HandleErr(err)
		sz = prettyByteSize(int(i.Size()))
		m[file.Name()] = sz
		PublicFiles = append(PublicFiles, file.Name(), sz)
	}

	//return PublicFiles
	return m
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {

	appVersion := "0.2.0"
	hostname, err := os.Hostname()
	util.HandleErr(err)
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
		util.GetTime(),
		hostname,
		publicFiles,
		HostsPingStat,
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
	_, err := responseOutput(w, time.Now().Format("02 Jan 2006 15:04:05 MST"))
	util.HandleErr(err)
}

// Main logic
func main() {

	// Get config and determine location

	var dirStatus = strings.Contains(util.GetWorkDir(), ".")
	var wg = new(sync.WaitGroup)
	wg.Add(4)

	// Get arguments
	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&conf.CONFIG, "config", "config.yml", "Define config file")
	flag.Parse()
	if util.IsFlagPassed("config") {
		fmt.Println(`Argument "-config" passed`)
	}

	config, _ := conf.LoadConfig(conf.CONFIG, dirStatus)

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
			util.SigtermHandler(s)
		}
	}()

	exitcode := <-exitchnl
	os.Exit(exitcode)

	wg.Wait()
	// Routines end

}
