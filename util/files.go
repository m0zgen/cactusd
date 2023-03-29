package util

import (
	"bufio"
	"fmt"
	fileMerger "github.com/Ja7ad/goMerge"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Downloads

// Download - URL iterator
func Download(url []string, dest string) {
	//fmt.Println(url[1])
	for i, u := range url {
		fmt.Println(i, u)
		err := downloadFile(u, dest)
		HandleErr(err)
	}
}

// URL file downloader
func downloadFile(url string, dest string) error {
	var postfix = "_prev"
	var filename = GetFilenameFromUrl(url)
	var filepath = filepath.Join(dest, filename)
	if !strings.Contains(filename, ".txt") {
		filepath = filepath + ".txt"
	}

	// Check exists file for processing in future
	//if exists := getFileExists(filepath); exists == true {
	//	fmt.Printf("File exists %s\n", filename)
	//}

	exist := IsFileExists(filepath)
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

	mergedFileName := MergedDir + "/" + GetFilenameFromUrl(dest) + ".txt"
	if exist {
		matched, _ := IsFileMatched(filepath, filepath+postfix)
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

// MergeFiles - Merge downloaded files to one from folder to target
func MergeFiles(path string, ext string, dest string) {
	err := fileMerger.Merge(path, ext, dest, false)
	if err != nil {
		log.Fatal(err)
	}
}

// PublishFiles - Process merged file
func PublishFiles(mergeddir string, out string) {
	// Process merged files
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

		SortFile(f)

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

	if IsFileExists(dst) {
		DeleteFile(dst)
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

// SortFile - Sort and remove duplicates from files
func SortFile(file string) {

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
