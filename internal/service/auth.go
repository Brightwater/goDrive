package service

import (
	"fmt"
	"log"
	"net/http"
)

// call the oauth service
func VerifyTokenAndScope(token string) error {

    url := AppConfig.AuthBaseURL + "/verifyTokenAndScope?token=" + token + "&scope=goDrive"

	log.Println(url)

    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        return err
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("unauthorized")
	}
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return nil
}