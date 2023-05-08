package file

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Brightwater/goDrive/internal/db"
	"github.com/Brightwater/goDrive/internal/service"
)


func CleanUploadedFiles() {

	path := service.AppConfig.UploadPath

	names, err := db.GetListOfUploadFiles()
	if err != nil {
		log.Printf("Error in getting db uploaded files %s", err)
	}

	files, err := filepath.Glob(filepath.Join(path, "*"))
    if err != nil {
        log.Println(err)
        return
    }

	// Loop over files and delete any file that is not in the list of filenames
    for _, file := range files {

        fileName := filepath.Base(file)
        
		if !contains(names, fileName) {
            err := os.Remove(file)
            if err != nil {
                log.Println(err)
            } else {
                log.Printf("Deleted uploaded file: %s\n", file)
            }
        }
    }
}

// Helper function to check if a string is present in a slice of strings
func contains(s []string, str string) bool {
    for _, v := range s {
        if v == str {
            return true
        }
    }
    return false
}