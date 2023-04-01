package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func MoveFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("sync error: %s", err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat error: %s", err)
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return fmt.Errorf("chmod error: %s", err)
	}

	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

// URL file downloader
func downloadFile(url string, saveFile string, dest string) error {

	var filename = GetFilenameFromUrl(url)

	// Create the file
	out, err := os.Create(saveFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	// TODO: detect 404 pages or 200 response
	fmt.Printf("Downloading file %s\n", saveFile)
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

	mergedFileName := MergedDir + "/" + GetFilenameFromUrl(dest) + ".txt"
	exist := IsFileExists(saveFile)

	if exist {
		matched, _ := IsFileMatched(saveFile, saveFile+"_prev")
		if matched {
			fmt.Println("Previous and current files - matched. No needed action.")
		} else {
			fmt.Printf("Merging files: %s\n", filename)

			if IsFileExists(mergedFileName) {
				DeleteFile(mergedFileName)
			}
			MergeFiles(dest, ".txt", mergedFileName)
		}
	} else {
		if IsFileExists(mergedFileName) {
			DeleteFile(mergedFileName)
		}
		MergeFiles(dest, ".txt", mergedFileName)
	}

	return nil
}

// Download - URL iterator
func Download(url []string, dest string) {
	//fmt.Println(url[1])
	var postfix = "_prev"

	for i, u := range url {
		fmt.Println(i, u)

		var dwnFileName = GetFilenameFromUrl(u)
		var saveFile = filepath.Join(dest, dwnFileName)

		if !strings.Contains(saveFile, ".txt") {
			saveFile = saveFile + ".txt"
		}

		exist := IsFileExists(saveFile)
		if exist {
			err := MoveFile(saveFile, saveFile+postfix)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				log.Fatal(err)
				//os.Exit(1)
			}
			//e := os.Rename(filepath, filepath+postfix)
			//if e != nil {
			//	log.Fatal(e)
			//}
		}

		err := downloadFile(u, saveFile, dest)
		HandleErr(err)
	}
}
