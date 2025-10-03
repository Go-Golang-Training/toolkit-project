package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Go-Golang-Training/toolkit"
)

func main() {
	mux := routes()

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}

}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/upload", uploadFiles)
	mux.HandleFunc("/upload-one", uploadOneFile)

	return mux
}

func uploadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	t := toolkit.Tools{
		MaxFileSize:      1024 * 1024 * 1024, // 1 GB
		AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
	}

	files, err := t.UploadFiles(r, "./uploads")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out := ""
	for _, item := range files {
		out += fmt.Sprintf("Uploaded %s to the uploads directory, renamed to %s and is %d bytes in size\n", item.OriginalFileName, item.NewFileName, item.FileSize)

	}

	_, _ = w.Write([]byte(out))
}

func uploadOneFile(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	t := toolkit.Tools{
		MaxFileSize:      1024 * 1024 * 1024, // 1 GB
		AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
	}

	file, err := t.UploadOneFile(r, "./uploads")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out := ""

	out += fmt.Sprintf("Uploaded %s to the uploads directory, renamed to %s and is %d bytes in size\n", file.OriginalFileName, file.NewFileName, file.FileSize)

	_, _ = w.Write([]byte(out))

}
