package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func createImageUrl(dockerRegistryUrl, image string) (string, error) {
	i := strings.Split(image, ":")
	if len(i) < 2 {
		return "", fmt.Errorf("No tag provide for the input reference \"%s\" reference, please provide one.", image)
	}
	tag := i[1]
	ref := i[0]

	arr := []string{dockerRegistryUrl, ref, "manifests", tag}
	return strings.Join(arr, "/"), nil
}

func getManifest(dockerRegistryUrl, image string) (Manifest, error) {
	var manifest Manifest
	imageUrl, err := createImageUrl(dockerRegistryUrl, image)
	if err != nil {
		return manifest, err
	}

	resp, _ := http.Get(imageUrl)

	if resp.StatusCode == http.StatusUnauthorized && token == "" {
		authHeader := resp.Header["Www-Authenticate"][0]
		token = getToken(getAuthParams(authHeader))
	}

	req, _ := http.NewRequest("GET", imageUrl, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	var client http.Client
	resp, err = client.Do(req)
	if err != nil {
		return manifest, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return manifest, err
	}

	if resp.StatusCode >= 400 || resp.StatusCode < 200 {
		return manifest, fmt.Errorf("Impossible to get the manifest, got error status code from the registry: %d \nErrorBody: %s", resp.StatusCode, string(body))
	}

	json.Unmarshal(body, &manifest)

	return manifest, nil
}

func GetManifest(dockerRegistryUrl, image string) (Manifest, error) {
	manifest, err := getManifest(dockerRegistryUrl, image)
	return manifest, err
}
