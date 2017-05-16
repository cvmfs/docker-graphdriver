package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	cvmfsUtil "github.com/cvmfs/docker-graphdriver/plugins/util"
)

const (
	dockerAuthUrl     = "https://auth.docker.io"
	dockerRegistryUrl = "https://registry-1.docker.io/v2"
	thinImageVersion  = "0.1"
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

func getBlob(url string) []byte {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	var client http.Client
	resp, _ := client.Do(req)

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return body
}

func getConfig(url string) []byte {
	return getBlob(url)
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

func PullLayers(args []string) {
	if len(args) != 1 {
		printUsage()
		return
	}

	image := args[0]
	manifest := getManifest(image)

	os.Mkdir("/tmp/layers", 0755)
	for idx, layer := range manifest.Layers {

		fmt.Printf("%2d: %s\n", idx, layer.Digest)
		getLayer(image, layer.Digest)
	}
}

func GetManifest(args []string) (Manifest, error) {
	var manifest Manifest

	if len(args) != 1 {
		printUsage()
		return manifest, errors.New("Not enough arguments...")
	} else {
		manifest = getManifest(args[0])
		return manifest, nil
	}
}

func GetConfig(args []string) {
	manifest, _ := GetManifest(args)
	repo := args[0]

	config_digest := manifest.Config.Digest
	url := dockerRegistryUrl + "/" + repo + "/blobs/" + config_digest

	resp := getConfig(url)

	fmt.Println(string(resp))
}

func MakeThinImage(m Manifest, repo string) cvmfsUtil.ThinImage {
	layers := make([]cvmfsUtil.ThinImageLayer, len(m.Layers))

	for i, l := range m.Layers {
		d := strings.Split(l.Digest, ":")[1]
		layers[i] = cvmfsUtil.ThinImageLayer{Digest: d, Repo: repo}
	}

	return cvmfsUtil.ThinImage{Layers: layers, Version: thinImageVersion}
}
