package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type UploadHandler struct {
	HostAddr  string
	UploadDir string
	Filter    string
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		filterValue := r.FormValue(h.Filter)
		//fmt.Fprintf(w, "Parsed query-param with key \"%s\": %s", h.Filter, filterValue)
		err := filepath.Walk(h.UploadDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				ext := filepath.Ext(path)
				if filterValue == "" || ext == filterValue {
					fmt.Fprintf(w, "name: %s / extension: %s / size: %d\n", info.Name(), ext, info.Size())
				}
			}
			return nil
		})
		if err != nil {
			http.Error(w, "Unable to get file list", http.StatusInternalServerError)
		}
	case http.MethodPost:
		fmt.Println("post")
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		files, err := os.ReadDir(h.UploadDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if f.Name() == header.Filename {
				fmt.Fprintf(w, "File %s already exists\n", header.Filename)
				return
			}
		}

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusBadRequest)
			return
		}

		filePath := h.UploadDir + "/" + header.Filename
		err = os.WriteFile(filePath, data, 0777)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "File %s has been successfully uploaded\n", header.Filename)

		fileLink := h.HostAddr + "/" + header.Filename
		fmt.Fprintln(w, fileLink)
	}
}

func main() {
	uploadHandler := &UploadHandler{
		HostAddr:  "http://localhost:8080",
		UploadDir: "/home/slava/go/src/projects/Go-Backend-Level-1/lesson-4/upload",
		Filter:    "ext",
	}
	http.Handle("/upload", uploadHandler)
	http.Handle("/", uploadHandler)

	go func() {
		dirToServe := http.Dir(uploadHandler.UploadDir)
		fs := &http.Server{
			Addr:         ":8080",
			Handler:      http.FileServer(dirToServe),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		log.Println("file server start")
		err := fs.ListenAndServe()
		if err != nil {
			log.Fatal("Server ListenAndServe: ", err)
		}
	}()

	srv := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("server start")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Server ListenAndServe: ", err)
	}
}
