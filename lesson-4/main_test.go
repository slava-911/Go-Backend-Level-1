package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8081/?ext=.png", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	uploadHandler := &UploadHandler{
		HostAddr:  "http://localhost:8080",
		UploadDir: "/home/slava/go/src/projects/Go-Backend-Level-1/lesson-4/upload",
		Filter:    "ext",
	}

	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "name: testfile5.png / extension: .png / size: 0\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestUploadHandler(t *testing.T) {
	file, _ := os.Open("testfile")
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()

	uploadHandler := &UploadHandler{
		UploadDir: "/home/slava/go/src/projects/Go-Backend-Level-1/lesson-4/upload",
		HostAddr:  ts.URL,
	}

	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := `testfile`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
