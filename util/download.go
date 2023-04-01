package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// URL file downloader
func downloadFile(url string, saveFile string, dest string) error {

	//var filename = GetFilenameFromUrl(url)
	//var prevFile = saveFile + PrevPrefix

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

	// Generate out file name from destination catalog name
	mergedFileName := MergedDir + "/" + GetFilenameFromUrl(dest) + ".txt"

	if IsFileExists(mergedFileName) {
		DeleteFile(mergedFileName)
	}

	fmt.Println("AAAAAAAAAAAAAAAAAAAAA", dest, mergedFileName)
	MergeFiles(dest, ".txt", mergedFileName)

	//exist := IsFileExists(prevFile)
	//
	//if exist {
	//	matched, _ := IsFileMatched(saveFile, prevFile)
	//	if matched {
	//		fmt.Println("Previous and current files - matched. No needed action.")
	//	} else {
	//		fmt.Printf("Merging files: %s\n", filename)
	//
	//		if IsFileExists(mergedFileName) {
	//			DeleteFile(mergedFileName)
	//		}
	//
	//		fmt.Println("AAAAAA", dest)
	//		MergeFiles(dest, ".txt", mergedFileName)
	//	}
	//} else {
	//	if IsFileExists(mergedFileName) {
	//		DeleteFile(mergedFileName)
	//	}
	//	MergeFiles(dest, ".txt", mergedFileName)
	//}

	return nil
}

// Download - URL iterator
func Download(url []string, dest string) {
	//fmt.Println(url[1])
	for i, u := range url {
		fmt.Println(i, u)

		var dwnFileName = GetFilenameFromUrl(u)
		var saveFile = filepath.Join(dest, dwnFileName)

		if !strings.Contains(saveFile, ".txt") {
			saveFile = saveFile + ".txt"
		}

		exist := IsFileExists(saveFile)
		if exist {
			err := MoveFile(saveFile, saveFile+PrevPrefix)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				//log.Fatal(err)
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
