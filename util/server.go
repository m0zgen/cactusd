package util

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

// Web server handlers

func responseOutput(w http.ResponseWriter, message string) (int, error) {
	return fmt.Fprint(w, message)
}

// Pretty bite
// Thx: https://stackoverflow.com/questions/1094841/get-human-readable-version-of-file-size/1094933#1094933
func prettyByteSize(b int) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

// Walk file public names
func listPublicFilesDir(target string) map[string]string {
	files, err := os.ReadDir(target)
	m := make(map[string]string)
	var sz string
	var PublicFiles []string
	if err != nil {
		HandleErr(err)
	}

	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())
		//fmt.Println(target + file.Name())
		i, err := file.Info()
		HandleErr(err)
		sz = prettyByteSize(int(i.Size()))
		m[file.Name()] = sz
		PublicFiles = append(PublicFiles, file.Name(), sz)
	}

	//return PublicFiles
	return m
}

// TimeHandler - Show rime
func TimeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := responseOutput(w, time.Now().Format("02 Jan 2006 15:04:05 MST"))
	HandleErr(err)
}

// File uploader
func webUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file") // the FormFile function takes in the POST input id file

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	//

	f, err := os.OpenFile("./upload/"+header.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	HandleErr(err)

	defer f.Close()
	_, err = io.Copy(f, file)

	_, err = responseOutput(w, "File uploaded successfully : ")
	_, err = responseOutput(w, header.Filename)
	HandleErr(err)

}

// Template bind
func serveTemplate(w http.ResponseWriter, r *http.Request) {

	appVersion := AppVersion
	hostname, err := os.Hostname()
	HandleErr(err)
	publicFiles := listPublicFilesDir("./public/files/")

	//
	files := []string{
		"./templates/base.html",
		"./templates/partials/nav.html",
		"./templates/home.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	////
	data := struct {
		AppVersion  string
		CurrentDate string
		HostName    string
		PublicFiles map[string]string
		PingStatus  map[string]string
	}{
		appVersion,
		GetTime(),
		hostname,
		publicFiles,
		HostsPingStat,
	}

	//err = ts.Execute(w, data)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//}
	////

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}

}

// RunHttpServer - Start HTTP sever
func RunHttpServer(port string) {

	//fileHandler := http.StripPrefix("/", http.FileServer(http.Dir("public")))

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", serveTemplate)

	//http.Handle("/", fileHandler)
	http.HandleFunc("/time", TimeHandler)
	http.HandleFunc("/upload", webUpload)

	// Run server
	//handler := http.FileServer(http.Dir("./public"))
	//http.Handle("/download", handler)

	log.Print("Listening on : " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
