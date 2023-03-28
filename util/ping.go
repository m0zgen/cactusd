package util

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var hostsPingStat = make(map[string]string)

type PingParams struct {
	name string
	port int
}

// PingHost Ping target
func PingHost(host string, p int) {
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

// CallPinger PingHost caller from unmarshal config wit ping params
func CallPinger() {

	var dirStatus = strings.Contains(GetWorkDir(), ".")

	//configData := LoadUnmarshalConfig(CONFIG, dirStatus)
	InitYConfig(CONFIG, dirStatus)
	configData := GetYConfig()

	pingConfig := configData["ping"].([]interface{})
	var p PingParams
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
		PingHost(p.name, p.port)
	}
}
