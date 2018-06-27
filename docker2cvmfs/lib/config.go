package lib

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func getBlob(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in making the request to the registry.")
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error in reading the request.")
		return nil, err
	}
	resp.Body.Close()

	return body, nil
}

func getConfig(url string) ([]byte, error) {
	return getBlob(url)
}

func GetConfig(dockerRegistryUrl string, imageReference string) (string, error) {
	manifest, _ := GetManifest(dockerRegistryUrl, imageReference)
	repo := strings.Split(imageReference, ":")[0]

	config_digest := manifest.Config.Digest
	url := dockerRegistryUrl + "/" + repo + "/blobs/" + config_digest

	resp, err := getConfig(url)

	return string(resp), err
}
