package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func PullLayers(dockerRegistryUrl string, args []string) {
	if len(args) != 1 {
		printUsage()
		return
	}

	reference := args[0]
	manifest := getManifest(dockerRegistryUrl, reference)

	image := strings.Split(reference, ":")[0]

	os.Mkdir("/tmp/layers", 0755)
	for idx, layer := range manifest.Layers {
		fmt.Printf("%2d: %s\n", idx, layer.Digest)
		getLayer(dockerRegistryUrl, image, layer.Digest)
	}
}

func getLayer(dockerRegistryUrl, repo, digest string) {
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
