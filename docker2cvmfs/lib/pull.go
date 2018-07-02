package lib

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	//"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func PullLayers(dockerRegistryUrl, inputReference, repository, subdirectory string) error {
	manifest := getManifest(dockerRegistryUrl, inputReference)
	image := strings.Split(inputReference, ":")[0]

	destDir, err := ioutil.TempDir("", "docker2cvmfs_layers")
	if err != nil {
		return err
	}

	for _, layer := range manifest.Layers {
		// TODO make use of cvmfsRepo and cvmfsSubDirectory
		err := getLayer(dockerRegistryUrl, image, layer.Digest, destDir, repository, subdirectory)
		if err != nil {
			return err
		}
	}
	return nil
}

func MakeRequestToRegistry(dockerRegistryUrl, repository, digest string) (*http.Response, error) {
	url := dockerRegistryUrl + "/" + repository + "/blobs/" + digest
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	var client http.Client
	return client.Do(req)
}

func getLayer(dockerRegistryUrl, repository, digest, destDir, cvmfsRepo, cvmfsSubDirectory string) error {
	hash := strings.Split(digest, ":")[1]
	filename := hash + ".tar.gz"

	file, err := os.Create(destDir + "/" + filename)
	if err != nil {
		fmt.Println("Impossible to create the file: ", filename, "\nerr: ", err)
		return err
	}

	resp, err := MakeRequestToRegistry(dockerRegistryUrl, repository, digest)

	if err != nil {
		fmt.Println(err)
		return err
	}

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	_, err = io.Copy(file, tee)
	if err != nil {
		fmt.Println(err)
		return err
	}

	uncompressed, err := gzip.NewReader(&buf)
	if err != nil {
		fmt.Println(err)
		return err
	}
	cmd := exec.Command("cvmfs_server", "ingest", "-t", "-", "-b", cvmfsSubDirectory+hash, cvmfsRepo)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()

	go func() {
		_, err := io.Copy(stdin, uncompressed)
		if err != nil {
			fmt.Println("Error in writing to stdin: ", err)
		}
	}()

	if err != nil {
		return err
	}

	slurpOut, err := ioutil.ReadAll(stdout)

	slurpErr, err := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	fmt.Println("STDOUT: ", string(slurpOut))
	fmt.Println("STDERR: ", string(slurpErr))
	fmt.Println("Wrote file to: ", filename)
	file.Close()
	resp.Body.Close()
	return nil
}
