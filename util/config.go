package util

import (
	"gopkg.in/yaml.v3"
	"os"
)

const MergedDir string = "merged"

var CONFIG string
var HostsPingStat = hostsPingStat

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

// LoadConfig Config file loader
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

func loadUnmarshalConfig(filename string, dirStatus bool) map[string]interface{} {

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
