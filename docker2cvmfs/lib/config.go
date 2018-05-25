package lib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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

func GetConfig(dockerRegistryUrl string, args []string) {
	manifest, _ := GetManifest(dockerRegistryUrl, args[0])
	repo := strings.Split(args[0], ":")[0]

	config_digest := manifest.Config.Digest
	url := dockerRegistryUrl + "/" + repo + "/blobs/" + config_digest

	resp := getConfig(url)

	fmt.Println(string(resp))
}
