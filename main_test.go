package main

import (
	"cactusd/util"
	"testing"
)

// Test downloadFile function
func TestDownloadFile(t *testing.T) {
	url := "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
	saveFile := "hosts.txt"
	dest := "/tmp"
	err := util.DownloadFile(url, saveFile, dest)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}
