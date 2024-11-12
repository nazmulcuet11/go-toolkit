package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nazmulcuet11/go-toolkit/toolkit"
)

func main() {
	mux := routes()
	log.Println("Starting server on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/upload", uploadFiles)
	mux.HandleFunc("/upload-one", uploadFile)

	return mux
}

func uploadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	t := toolkit.Tools{}
	t.MaxFileSize = 1 << 30 // I GB
	t.AllowedFileTypes = []string{"image/jpeg", "image/png", "image/gif"}
	files, err := t.UploadFiles(r, "./uploads")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	out := ""
	for _, file := range files {
		out += fmt.Sprintf("Uploaded %s, to /uploads folder, renamed to %s\n", file.OriginalFileName, file.NewFileName)
	}

	_, _ = w.Write([]byte(out))
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	t := toolkit.Tools{}
	t.MaxFileSize = 1 << 30 // I GB
	t.AllowedFileTypes = []string{"image/jpeg", "image/png", "image/gif"}
	file, err := t.UploadFile(r, "./uploads")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	out := fmt.Sprintf("Uploaded %s, to /uploads folder, renamed to %s\n", file.OriginalFileName, file.NewFileName)

	_, _ = w.Write([]byte(out))
}
