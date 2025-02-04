package internal

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Brightwater/goDrive/internal/auth"
	"github.com/Brightwater/goDrive/internal/db"
	"github.com/Brightwater/goDrive/internal/service"
	"github.com/go-chi/chi/v5"
)

var charset = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	applyMiddleWare(r)

	// r.Handle("/*", http.FileServer(http.Dir("static/index.html")))

	r.Get("/hello", testGet)
	r.Get("/authTest", testGetWithAuth)
	r.Get("/form", getForm)
	r.Get("/permissionCode/{hours}/{uses}", getPermissionCode)
	r.Get("/downloadFile/{fileName}/{code}", downloadHandler)

	r.Post("/localFileAdd/{hours}", addLocalFileDownload)
	r.Post("/upload", uploadHandler)

	return r
}

func testGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

func getPermissionCode(w http.ResponseWriter, r *http.Request) {
	err := auth.VerifyTokenAndScope(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	hours, err := strconv.ParseInt(chi.URLParam(r, "hours"), 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uses, err := strconv.ParseInt(chi.URLParam(r, "uses"), 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	code, err := createCode()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// persist the code
	err = db.PersistPermissionCode(code, hours, uses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(code))
}

func createCode() (string, error) {
	length := 6

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[int(b)%len(charset)]
	}

	code := string(bytes)
	return code, nil
}

func testGetWithAuth(w http.ResponseWriter, r *http.Request) {
	err := auth.VerifyAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "Hello world")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "fileName")
	code := chi.URLParam(r, "code")

	var filePath string
	localFile, err := db.GetLocalFileDownload(fileName, code)

	if err != nil {
		log.Println("Local file not found, checking upload files")
		upFile, err := db.GetUploadedFile(fileName, code)
		if err != nil {
			http.Error(w, "Failed to get file", http.StatusInternalServerError)
			return
		}
		filePath = service.AppConfig.UploadPath + "/" + upFile.FileName
	} else {
		filePath = localFile.FilePath
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file info: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	mode := fileInfo.Mode()
	if mode.IsDir() {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName+".zip"))

		log.Println("Path is a directory")

		buf, err := handleADirectory(w, filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the zip archive content to the HTTP response writer
		_, err = w.Write(buf.Bytes())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if mode.IsRegular() {
		log.Println("Path is a file")

		file, err := os.Open(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		// Set the Content-Disposition header to prompt the browser to download the file
    	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

		// Serve the file content to the client
		http.ServeContent(w, r, fileName, fileInfo.ModTime(), file)

	} else {
		http.Error(w, fmt.Sprintf("Failed to get file info: %s", err.Error()), http.StatusInternalServerError)
	}
}

func handleADirectory(w http.ResponseWriter, filePath string) (*bytes.Buffer, error) {
	w.Header().Set("Content-Type", "application/zip")

	zipWriter := zip.NewWriter(w)

	// Create a new in-memory buffer to store the zip archive
	buf := new(bytes.Buffer)

	// Walk the directory tree and add each file to the zip archive
	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Open the file to be added to the zip archive
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Create a new zip file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set the name of the file to be added to the zip archive
		header.Name = filepath.Base(path)

		// Add the file to the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy the file content to the zip archive writer
		if _, err := io.Copy(writer, file); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	log.Println("All files added to zip archive successfully")

	// Close the zip archive writer to flush any remaining data to the HTTP response writer
	err = zipWriter.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return buf, nil
}

type JsonData struct {
	FilePath string `json:"filePath"`
}

// path, hours
func addLocalFileDownload(w http.ResponseWriter, r *http.Request) {

	err := auth.VerifyTokenAndScope(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	hours, err := strconv.ParseInt(chi.URLParam(r, "hours"), 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if hours == 0 {
		hours = 318000
	}

	var jsonData JsonData
	err = json.NewDecoder(r.Body).Decode(&jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filePath := jsonData.FilePath

	fileName := filepath.Base(filePath)

	code, err := createCode()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.PersistLocalFileDl(filePath, fileName, code, hours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "/downloadFile/%s/%s", fileName, code)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	code := r.Header.Get("Authorization")
	if strings.Contains(code, "Bearer") {
		http.Error(w, "error", http.StatusUnauthorized)
		return
	}

	err := auth.VerifyPermissionCode(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Print("Uploading file")

	// Get the boundary from the request Content-Type header
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Print("Uploading file 2")

	// Get the boundary from the media type
	boundary := params["boundary"]
	mr := multipart.NewReader(r.Body, boundary)

	fmt.Print("Uploading file 3")

	// Get the file part
	fh, err := mr.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Print("Uploading file 4")

	uploadPath := service.AppConfig.UploadPath + "/"

	// Create a new file in the specified directory
	f, err := os.Create(uploadPath + fh.FileName())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fmt.Print("Uploading file 5")

	// Copy the file contents to the output file
	io.Copy(f, fh)

	err = db.UpdatePermissionCodeUsesCount(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Print("Uploading file 6")

	ip := r.RemoteAddr

	err = db.PersistUploadFile(fh.FileName(), code, ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Print("Uploading file 7")

	// Return a success message to the client
	fmt.Fprintf(w, "/downloadFile/%s/%s", fh.FileName(), code)

}

func getForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "base.html")
}
