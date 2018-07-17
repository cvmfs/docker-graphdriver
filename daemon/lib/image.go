package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type ManifestRequest struct {
	Anonymous bool
	Password  <-chan string
	Token     <-chan string
}

func (img Image) GetManifest(request ManifestRequest) ([]byte, error) {
	if request.Anonymous {
		return img.getAnonymousManifest()
	} else {
		select {
		case password := <-request.Password:
			fmt.Println("Got password")
			return img.getManifestWithPassword(password)
		case token := <-request.Token:
			return img.getManifestWithRefreshToken(token)
		case <-time.After(1 * time.Second):
			LogE(nil).Warning("Didn't receive neither password nor token, trying to get the manifest anonymously")
			return img.getAnonymousManifest()
		}
	}
}

func (img Image) getAnonymousManifest() ([]byte, error) {
	return getManifestWithUsernameAndPassword(img, "", "")
}

func (img Image) getManifestWithPassword(password string) ([]byte, error) {
	return getManifestWithUsernameAndPassword(img, img.User, password)
}

func getManifestWithUsernameAndPassword(img Image, user, pass string) ([]byte, error) {

	url := img.GetManifestUrl()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 401 {
		log.WithFields(log.Fields{
			"status code": resp.StatusCode,
		}).Info("Expected status code 401, print body anyway.")
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			LogE(err).Error("Error in reading the first http response")
		}
		fmt.Println(string(body))
		return nil, err
	}
	WwwAuthenticate := resp.Header["Www-Authenticate"][0]
	token, _, err := requestAuthToken(WwwAuthenticate, user, pass)
	if err != nil {
		LogE(err).Error("Error in getting the authentication token")
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		LogE(err).Error("Impossible to create a http request")
		return nil, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err = client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogE(err).Error("Error in reading the second http response")
		return nil, err
	}
	return body, nil
}

func (img Image) getManifestWithRefreshToken(token string) ([]byte, error) {
	return nil, nil
}

func parseBearerToken(token string) (realm string, options map[string]string, err error) {
	options = make(map[string]string)
	args := token[7:]
	keyValue := strings.Split(args, ",")
	for _, kv := range keyValue {
		splitted := strings.Split(kv, "=")
		if len(splitted) != 2 {
			err = fmt.Errorf("Wrong formatting of the token")
			return
		}
		splitted[1] = strings.Trim(splitted[1], `"`)
		if splitted[0] == "realm" {
			realm = splitted[1]
		} else {
			options[splitted[0]] = splitted[1]
		}
	}
	return
}

func requestAuthToken(token, user, pass string) (authToken, refreshToken string, err error) {
	realm, options, err := parseBearerToken(token)
	if err != nil {
		return
	}
	req, err := http.NewRequest("GET", realm, nil)
	if err != nil {
		return
	}

	query := req.URL.Query()
	for k, v := range options {
		query.Add(k, v)
	}
	if user != "" && pass != "" {
		query.Add("offline_token", "true")
		req.SetBasicAuth(user, pass)
	}
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = fmt.Errorf("Authorization error %s", resp.Status)
		return
	}

	var jsonResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	authTokenInterface, ok := jsonResp["token"]
	if ok {
		authToken = "Bearer " + authTokenInterface.(string)
	} else {
		err = fmt.Errorf("Didn't get the token key from the server")
		return
	}
	if user != "" && pass != "" {
		refreshTokenInterface, ok := jsonResp["refresh_token"]
		if ok {
			refreshToken = refreshTokenInterface.(string)
		} else {
			err = fmt.Errorf("Didn't get the refresh token from the server")
			return
		}
	}
	return
}
