package util

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// SigtermHandler catch Ctrl+C or SIGTERM
func SigtermHandler(signal os.Signal) {
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

// HandleErr Error handler
func HandleErr(e error) {
	if e != nil {
		//panic(e)
		log.Println(e)
	}
}

// UpdatePath auto config file path updater
func UpdatePath(filename string) string {
	var path string
	path = GetWorkDir()
	filename = path + "/" + filename
	return filename
}

// GetWorkDir detect runner from binary or from "go run"
func GetWorkDir() string {
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

// GetTime return current time date with described format
func GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetFilenameFromUrl return last octet in passed string
// Thx: https://github.com/peeyushsrj/golang-snippets
func GetFilenameFromUrl(urlstr string) string {
	u, err := url.Parse(urlstr)
	if err != nil {
		log.Fatal("Error due to parsing url: ", err)
	}
	x, _ := url.QueryUnescape(u.EscapedPath())
	return filepath.Base(x)
}

// IsFlagPassed checks passed arguments for cactusd
func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// IsDirEmpty return bool value for caller
func IsDirEmpty(name string) bool {
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

// IsFileExists return bool value for caller
func IsFileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	} else {
		return false
	}
}

// IsFileMatched checks if two file is the same
func IsFileMatched(path1, path2 string) (sameSize bool, err error) {
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

// CreateDir create catalog in target place
func CreateDir(dirName string, dirStatus bool) error {

	if !dirStatus {
		dirName = UpdatePath(dirName)
	}
	err := os.MkdirAll(dirName, os.ModeSticky|os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

// DeleteFile delete target file
func DeleteFile(file string) {
	e := os.Remove(file)
	if e != nil {
		log.Fatal(e)
	}
}
