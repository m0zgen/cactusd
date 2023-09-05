package main

import (
	"cactusd/util"
	"testing"
)

// Test util.CreateDir function
func TestCreateDir(t *testing.T) {
	dirName := "/tmp/test"
	dirStatus := util.IsDirEmpty(dirName)
	err := util.CreateDir(dirName, dirStatus)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}

// Test util.downloadFile function
func TestDownloadFile(t *testing.T) {
	url := "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
	saveFile := "hosts.txt"
	dest := "/tmp"
	err := util.DownloadFile(url, saveFile, dest)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}
