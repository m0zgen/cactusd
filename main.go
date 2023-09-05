package main

import (
	"cactusd/util"
	conf "cactusd/util"
	"embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed templates/*
var templateFs embed.FS

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

// Clean stage on prepare step
// Ref: https://gist.github.com/m0zgen/af44035db3102d08effc2d38e56c01f3
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
		//`(?m)(^[$&+,:;=?@#|'<>.\-^*()%!].+$)|(^.*::.*)|(#|\s#.*)|(^.*\/\/.*)|(^.*,.*$)|(^.*\.-.*$)|(^.*[\$\^].*$)`)
		`(?m)(^[$&+,:;=?@#|'<>.\-^*()%!].+$)|(^.*::.*)|(#|\s#.*)|(^.*\/\/.*)|(^.*\.-.*$)|(^[а-я].*[--].*$)|(^[a-z].*[\^\,].*$)`)

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
	util.Download(config.Lists.Bl, config.Server.DownloadDir+"/bl")

	err = util.CreateDir(config.Server.DownloadDir+"/wl", dirStatus)
	util.Download(config.Lists.Wl, config.Server.DownloadDir+"/wl")

	err = util.CreateDir(config.Server.DownloadDir+"/bl_plain", dirStatus)
	util.Download(config.Lists.BlPlain, config.Server.DownloadDir+"/bl_plain")

	err = util.CreateDir(config.Server.DownloadDir+"/wl_plain", dirStatus)
	util.Download(config.Lists.WlPlain, config.Server.DownloadDir+"/wl_plain")

	err = util.CreateDir(config.Server.DownloadDir+"/ip_plain", dirStatus)
	util.Download(config.Lists.IpPlain, config.Server.DownloadDir+"/ip_plain")

	// Cleaning Process
	err = filepath.Walk(MergedDir, prepareFiles)
	util.HandleErr(err)
	util.PublishFiles(MergedDir, publishFilesDir)
	//
	if !util.IsDirEmpty(config.Server.UploadDir) {
		err := filepath.Walk(config.Server.UploadDir, prepareFiles)
		util.HandleErr(err)

		mergedFileName := publishFilesDir + "/dropped_ip.txt"
		util.MergeFiles(config.Server.UploadDir, ".txt", mergedFileName)
		util.SortFile(mergedFileName)
	}

}

func runTicker(config conf.Config, dirStatus bool, group *sync.WaitGroup, onlyGenerate bool) {
	defer group.Done()

	initial(config, dirStatus)

	if onlyGenerate {
		fmt.Println("Generated files stored in public/files catalog. Exit. Bye.")
		os.Exit(0)
	}

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

// type PublicFiles

// Main logic
func main() {

	// Get config and determine location

	var dirStatus = strings.Contains(util.GetWorkDir(), ".")
	util.TemplateFs = templateFs

	var wg = new(sync.WaitGroup)
	wg.Add(4)

	// Get arguments
	//Add usage ./cactusd -config <config ath or name>
	flag.StringVar(&conf.CONFIG, "config", "config.yml", "Define config file")
	showVersion := flag.Bool("version", false, "Show Cactusd version")
	onlyGenerate := flag.Bool("generate", false, "Run only as file generator and exit")

	flag.Parse()
	if util.IsFlagPassed("config") {
		fmt.Println(`Argument "-config" passed: `, conf.CONFIG)
	}

	if *showVersion {
		fmt.Println("Cactusd Version: ", conf.AppVersion)
		return
	}

	if *onlyGenerate {
		fmt.Println("Generate files...")
		// run generator
		//os.Exit(0)
	}

	config, _ := conf.LoadConfig(conf.CONFIG, dirStatus)

	conf.InitYConfig(conf.CONFIG, dirStatus)

	//serverConfig := configData["server"].(map[string]interface{})
	//listsConfig := configData["lists"].(map[string]interface{})

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

	// TODO: Add queue tasks

	if !*onlyGenerate {
		go util.RunHttpServer(config.Server.Port)
	}

	go runTicker(config, dirStatus, wg, *onlyGenerate)

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
