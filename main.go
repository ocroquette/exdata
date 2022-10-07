package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Compile templates on start of the application
var templates = template.Must(template.ParseFiles("public/upload.html"))

var repoPath = "repo"
var repoTmpPath = filepath.Join(repoPath, "tmp")

// Display the named template
func display(writer http.ResponseWriter, page string, data interface{}) {
	templates.ExecuteTemplate(writer, page+".html", data)
}

func storeDirForChecksum(cs string) string {
	if len(cs) < 8 {
		log.Fatal("Checksum too short: " + cs)
	}
	part1 := cs[0:4]
	part2 := cs[5:9]
	return filepath.Join(part1, part2)
}

func uploadFile(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(10 * 1024 * 1024)

	// Get handler for filename, size and headers
	inFile, handler, err := request.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer inFile.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create temporary output file
	outFile, err := os.CreateTemp(repoTmpPath, "upload-")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	h := sha256.New()

	const chunkSize = 100 * 1024

	buf := make([]byte, chunkSize)
	for {
		var n int

		// Read chunk
		n, err := inFile.Read(buf)

		if err != nil && err != io.EOF {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if err == io.EOF {
			break
		}

		outFile.Write(buf[:n])
		h.Write(buf[:n])
	}

	outFile.Close()
	bs := h.Sum(nil)
	checksumString := hex.EncodeToString(h.Sum(nil))
	repoSubdirRelative := storeDirForChecksum(checksumString)
	repoSubdirAbsolute := filepath.Join(repoPath, repoSubdirRelative)
	os.MkdirAll(repoSubdirAbsolute, 0755)
	e := os.Rename(outFile.Name(), filepath.Join(repoSubdirAbsolute, checksumString))

	if e != nil {
		log.Panic(e)
		fmt.Fprintf(writer, "Failed to rename file\n")
	}

	fmt.Fprintf(writer, "Successfully Uploaded File\n")
	fmt.Fprintf(writer, "%x\n", bs)
}

func uploadHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		display(writer, "upload", nil)
	case "POST":
		uploadFile(writer, request)
	}
}

func downloadHandler(writer http.ResponseWriter, request *http.Request) {
	checksumString := strings.TrimPrefix(request.URL.Path, "/download/")

	repoSubdirRelative := storeDirForChecksum(checksumString)
	repoSubdirAbsolute := filepath.Join(repoPath, repoSubdirRelative)

	http.ServeFile(writer, request, filepath.Join(repoSubdirAbsolute, checksumString))
}

func createDirectories() {
	err := os.MkdirAll(repoTmpPath, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	createDirectories()

	http.HandleFunc("/upload", uploadHandler)

	http.HandleFunc("/download/", downloadHandler)

	//Listen on port 8080
	http.ListenAndServe(":8080", nil)
}
