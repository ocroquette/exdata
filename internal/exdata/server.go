package exdata

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	//go:embed assets/upload.html
	uploadHtml []byte
)

type Server struct {
	repo Repository
}

func (s Server) uploadFile(writer http.ResponseWriter, request *http.Request) {
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
	// filepath.Join(repoPath, "tmp")
	outFile, err := os.CreateTemp(s.repo.DirectoryForTemporaryFiles(), "upload-")
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
	repoSubdirRelative := s.repo.SubPathForChecksum(checksumString)
	repoSubdirAbsolute := filepath.Join(s.repo.BaseDir, repoSubdirRelative)
	os.MkdirAll(repoSubdirAbsolute, 0755)
	e := os.Rename(outFile.Name(), s.repo.FilePathForChecksum(checksumString))

	if e != nil {
		log.Panic(e)
		fmt.Fprintf(writer, "Failed to rename file\n")
	}

	fmt.Fprintf(writer, "Successfully uploaded File\n")
	fmt.Fprintf(writer, "%x\n", bs)
}

func (s Server) uploadHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		writer.Write(uploadHtml)
	case "POST":
		s.uploadFile(writer, request)
	}
}

func (s Server) downloadHandler(writer http.ResponseWriter, request *http.Request) {
	checksumString := strings.TrimPrefix(request.URL.Path, "/download/")

	http.ServeFile(writer, request, s.repo.FilePathForChecksum(checksumString))
}

// listenAddress: localhost:8080
// repoDir: path in the local file system
func (s Server) Start(listenAddress string, repoDir string) {
	s.repo = MakeRepository(repoDir)

	http.HandleFunc("/upload", s.uploadHandler)

	http.HandleFunc("/download/", s.downloadHandler)

	http.ListenAndServe(listenAddress, nil)
}
