package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	dockerAuthUrl     = "https://auth.docker.io"
	dockerRegistryUrl = "https://registry-1.docker.io/v2"
)

type TokenMessage struct {
	Token string
}

type ConfigType struct {
	MediaType string
	Size      int
	Digest    string
}

type Layer struct {
	MediaType string
	Size      int
	Digest    string
}

type Manifest struct {
	SchemaVersion int
	MediaType     string
	Config        ConfigType
	Layers        []Layer
}

var token string

func main() {
	if len(os.Args) != 2 {
		printUsage()
		return
	}

	image := os.Args[1]
	manifest := getManifest(image)

	os.Mkdir("/tmp/layers", 0755)
	for idx, layer := range manifest.Layers {
		fmt.Printf("%2d: %s\n", idx, layer.Digest)
		getLayer(image, layer.Digest)
	}
}

func printUsage() {
	fmt.Println("You need to specify a docker image to download!")
}

func getManifest(image string) Manifest {
	imageUrl := createImageUrl(image)

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

func getLayer(repo, digest string) {
	filename := "/tmp/layers/" + strings.Split(digest, ":")[1] + ".tar.gz"

	file, _ := os.Create(filename)
	fmt.Println(filename)

	url := dockerRegistryUrl + "/" + repo + "/blobs/" + digest
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Authorization", "Bearer "+token)

	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	file.Close()
	resp.Body.Close()
}

func getAuthParams(authHeader string) map[string]string {
	params := make(map[string]string)

	for _, v := range strings.Split(authHeader, ",") {
		s := strings.Split(v, "=")
		cutset := "\""

		if s[0] == "Bearer realm" {
			params["realm"] = strings.Trim(s[1], cutset)
		} else if s[0] == "service" {
			params["service"] = strings.Trim(s[1], cutset)
		} else if s[0] == "scope" {
			params["scope"] = strings.Trim(s[1], cutset)
		}
	}

	return params
}

func getToken(authParams map[string]string) string {
	tokenUrl := authParams["realm"] + "?service=" + authParams["service"] +
		"&scope=" + authParams["scope"]

	resp, _ := http.Get(tokenUrl)

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var m TokenMessage
	json.Unmarshal(body, &m)

	return m.Token
}

func createImageUrl(image string) string {
	s := strings.Split(image, "/")
	repo := s[0]
	name := strings.Join(s[1:], "/")

	arr := []string{dockerRegistryUrl, repo, name, "manifests", "latest"}
	return strings.Join(arr, "/")
}

func checkImageName(image string) bool {
	return false
}
