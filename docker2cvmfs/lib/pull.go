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

	destDir := "/home/simo/tmp/layers"

	os.Mkdir(destDir, 0755)
	for idx, layer := range manifest.Layers {
		fmt.Printf("%2d: %s\n", idx, layer.Digest)

		// TODO make use of cvmfsRepo and cvmfsSubDirectory
		err := getLayer(dockerRegistryUrl, image, layer.Digest, destDir, repository, subdirectory)
		if err != nil {
			return err
		}
	}
	return nil
}

func getLayer(dockerRegistryUrl, repository, digest, destDir, cvmfsRepo, cvmfsSubDirectory string) error {
	hash := strings.Split(digest, ":")[1]
	filename := hash + ".tar.gz"

	file, err := os.Create(destDir + "/" + filename)
	if err != nil {
		fmt.Println("Impossible to create the file: ", filename, "\nerr: ", err)
		return err
	}
	fmt.Println(filename)

	url := dockerRegistryUrl + "/" + repository + "/blobs/" + digest
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Authorization", "Bearer "+token)

	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return err
	}

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	nc, err := io.Copy(file, tee)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Duplicated ", nc, " bytes")

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
		//fmt.Println("About to copying to stdin")
		_, err := io.Copy(stdin, uncompressed)
		//fmt.Println("Finish copying to stdin")
		if err != nil {
			fmt.Println("Error in writing to stdin: ", err)
			//log.Fatal(err)
		}
		//fmt.Sprintln("Copied %d bytes in stdin", n)
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
