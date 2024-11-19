package main

import (
	"log"
	"net/http"

	"github.com/nazmulcuet11/go-toolkit/toolkit"
)

func main() {
	mux := routes()
	println("Listening at :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/download", downloadFile)
	return mux
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	println("Downloading file")
	t := toolkit.Tools{}
	t.DownloadStaticFile(w, r, "./files", "pic.jpg", "puppy.jpg")
}
