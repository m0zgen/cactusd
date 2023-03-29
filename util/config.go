package util

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const MergedDir string = "merged"

var CONFIG string
var HostsPingStat = hostsPingStat

var (
	yconfig     map[string]interface{}
	yconfigLock = new(sync.RWMutex)
)

var (
	AppVersion = "0.2.3"
)

// Config file structure type
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

// LoadConfig - Config file loader
func LoadConfig(filename string, dirStatus bool) (Config, error) {
	var config Config
	// Check go run or run binary
	if !dirStatus {
		filename = UpdatePath(filename)
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

// LoadUnmarshalConfig - Load yml ad interface
func LoadUnmarshalConfig(filename string, dirStatus bool) map[string]interface{} {

	// Check go run or run binary
	if !dirStatus {
		filename = UpdatePath(filename)
	}
	configFile, err := os.ReadFile(filename)
	//fmt.Println(filename)
	HandleErr(err)

	var data map[string]interface{}
	err = yaml.Unmarshal(configFile, &data)
	HandleErr(err)

	return data
}

// test fly reload
func loadYConfig(filename string, dirStatus bool, fail bool) {
	if !dirStatus {
		filename = UpdatePath(filename)
	}
	file, err := os.ReadFile(filename)
	HandleErr(err)

	var data map[string]interface{}
	err = yaml.Unmarshal(file, &data)
	HandleErr(err)

	yconfigLock.Lock()
	yconfig = data
	yconfigLock.Unlock()
}

func GetYConfig() map[string]interface{} {
	yconfigLock.RLock()
	defer yconfigLock.RUnlock()
	return yconfig
}

// InitYConfig go calls init on start
func InitYConfig(configName string, dirStatus bool) {
	loadYConfig(configName, dirStatus, true)
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGUSR2)
	go func() {
		for {
			<-s
			loadYConfig(configName, dirStatus, false)
			log.Println("Reloaded")
		}
	}()
}
