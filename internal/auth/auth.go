package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/Brightwater/goDrive/internal/db"
	"github.com/Brightwater/goDrive/internal/service"
)

var tokenCache = cache.New(15*time.Minute, 15*time.Minute)

// check the auth header and see if it contains either a permission code
// or a bearer token
// use either one to authenticate
func VerifyAuth(r *http.Request) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return fmt.Errorf("no token")
	}

	log.Printf("Auth header received: %s", token)

	if strings.Contains(token, "Bearer") {
		return VerifyTokenAndScope(r)
	}

	return VerifyPermissionCode(token)
}

func VerifyPermissionCode(token string) error {

	retCode, err := db.GetPermissionCode(token)
	if err != nil {
		log.Printf("couldn't get perm code %s", err)
		return err
	}

	str, _ := json.MarshalIndent(retCode, "", "\t")
	log.Printf("Retrieved permission code %s", str)

	if retCode.ExpirationTime.Before(time.Now()) || retCode.UsesRemaining == 0 {
		log.Println("Code expired")
		return fmt.Errorf("unauthorized")
	}

	return nil
}

// call the oauth service and check the token
func VerifyTokenAndScope(r *http.Request) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return fmt.Errorf("unauthorized")
	}

	token = strings.TrimPrefix(token, "Bearer ")

	_, found := tokenCache.Get(token)
	if found {
		log.Println("Token validated using cache")
		return nil
	}

	url := service.AppConfig.AuthBaseURL + "/verifyTokenAndScope?token=" + token + "&scope=goDrive"

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

	tokenCache.Set(token, token, cache.DefaultExpiration)

	return nil
}
