package internal

import (
	"fmt"
	"goDrive/internal/service"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

const maxRequestBodySize int64 = 2e+10 // 20Gb

func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	applyMiddleWare(r)

	r.Get("/hello", testGet)
	r.Get("/authTest/{token}", testGetWithAuth)
	r.Post("/upload", uploadHandler)
	r.Get("/form", getForm)

	return r
}

func testGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

func testGetWithAuth(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := service.VerifyTokenAndScope(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
        return
	}

	fmt.Fprintf(w, "Hello world")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	// Get the boundary from the request Content-Type header
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the boundary from the media type
	boundary := params["boundary"]
	mr := multipart.NewReader(r.Body, boundary)

	// Get the file part
	fh, err := mr.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new file in the specified directory
	f, err := os.Create("./uploadedFiles/" + fh.FileName())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Copy the file contents to the output file
	io.Copy(f, fh)

	// Return a success message to the client
	fmt.Fprintf(w, "File '%s' uploaded successfully!", fh.FileName())
}

func getForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "base.html")
}


