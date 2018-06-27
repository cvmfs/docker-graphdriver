package lib

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func createImageUrl(dockerRegistryUrl, image string) string {
	i := strings.Split(image, ":")
	tag := i[1]
	ref := i[0]

	arr := []string{dockerRegistryUrl, ref, "manifests", tag}
	return strings.Join(arr, "/")
}

func getManifest(dockerRegistryUrl, image string) Manifest {
	imageUrl := createImageUrl(dockerRegistryUrl, image)

	resp, _ := http.Get(imageUrl)

	if resp.StatusCode == http.StatusUnauthorized && token == "" {
		authHeader := resp.Header["Www-Authenticate"][0]
		token = getToken(getAuthParams(authHeader))
	}

	req, _ := http.NewRequest("GET", imageUrl, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	var client http.Client
	resp, _ = client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var manifest Manifest
	json.Unmarshal(body, &manifest)

	return manifest
}

func GetManifest(dockerRegistryUrl, image string) (Manifest, error) {
	manifest := getManifest(dockerRegistryUrl, image)
	return manifest, nil
}
