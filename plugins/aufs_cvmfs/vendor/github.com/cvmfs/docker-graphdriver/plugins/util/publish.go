package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"github.com/minio/minio-go"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
)

type MinioConfig struct {
	AccessKey        string
	AccessSecret     string
	Host             string
	CvmfsRepo        string
	SSL              bool
	PublishStatusURL string
}

var minioConfig MinioConfig

func readConfig() (config MinioConfig) {
	fmt.Println("reading minio config")

	var out []byte
	var err error

	if out, err = ioutil.ReadFile("/minio_ext_config/config.json"); err != nil {
		fmt.Println("Failed to read minio config.")
		fmt.Println(err)
		return
	}
	if err = json.Unmarshal(out, &config); err != nil {
		fmt.Println("Failed to parse minio config.")
		fmt.Println(err)
		return
	}

	fmt.Println("Minio config read successfully!")
	sc, _ := json.MarshalIndent(config, "", "    ")
	fmt.Println(string(sc))

	return config
}

func move(src string) (string, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	tmp := path.Join(os.TempDir(), fmt.Sprintf("dlcg-%d", rand.Int()))

	if err := os.MkdirAll(tmp, os.ModePerm); err != nil {
		fmt.Println("Failed to create tmp!")
		return "", err
	}

	var out bytes.Buffer
	cmd := fmt.Sprintf("mv %s/* %s", src, tmp)
	c := exec.Command("bash", "-c", cmd)
	c.Stdout = &out
	if err := c.Run(); err != nil {
		fmt.Println("Failed to move!")
		return "", err
	}

	return tmp, nil
}

func tar(src string) (string, error) {
	dstFile, err := ioutil.TempFile(os.TempDir(), "dlcg-tar-")
	defer dstFile.Close()

	if err != nil {
		fmt.Println("Failed to create temp file for tar.")
		return "", err
	}

	tarReader, err := archive.Tar(src, archive.Gzip)
	if err != nil {
		fmt.Println("Failed to create tar stream")
		return "", err
	}

	io.Copy(dstFile, tarReader)

	return dstFile.Name(), nil
}

func sha256hash(src string) (string, error) {
	tarFile, err := os.Open(src)
	defer tarFile.Close()

	if err != nil {
		fmt.Println("Failed to open file in hash()")
		return "", err
	}

	h := sha256.New()
	io.Copy(h, tarFile)
	s := fmt.Sprintf("%x", h.Sum(nil))

	return s, nil
}

func upload(src, h string) error {
	minioClient, err := minio.New(
		minioConfig.Host,
		minioConfig.AccessKey,
		minioConfig.AccessSecret,
		minioConfig.SSL)

	if err != nil {
		fmt.Println("Failed to create a minio client!")
		return err
	}

	for i := 0; i < 5; i++ {
		_, err = minioClient.FPutObject("layers", h, src, "application/x-gzip")
		if err != nil {
			fmt.Printf("Failed FPutObject(), this was attempt %d\n", i)
		} else {
			fmt.Printf("Upload successful, attempt %d\n!", i)
			return nil
		}
	}

	return fmt.Errorf("Failed to upload layer %s with hash %s\n", src, h)
}

func waitForPublishing(hash string) (err error) {
	for {
		target := minioConfig.PublishStatusURL + "/" + hash
		fmt.Println("waiting for publish.....")
		fmt.Println(target)

		client := http.Client{Timeout: time.Duration(2 * time.Second)}
		resp, err := client.Get(target)
		fmt.Println("got response")

		if err != nil {
			fmt.Println(err)
			return err
		}

		defer resp.Body.Close()
		buf, err := ioutil.ReadAll(resp.Body)
		body := string(buf)

		if resp.StatusCode != 200 {
			fmt.Println("Request failed, abort.")
			fmt.Println(body)
			return fmt.Errorf("status request failed, abort.")
		}

		if body == "publishing" {
			fmt.Println("Still publishing...")
			time.Sleep(1 * time.Second)
		} else if body == "done" {
			fmt.Println("Publishing done!")
			return nil
		} else if body == "unknown" {
			fmt.Println("Unknown publish status, abort.")
			return fmt.Errorf("Unknown publish status, abort.")
		} else {
			fmt.Println("Unknown reponse, abort.")
			fmt.Println(body)
			return fmt.Errorf("Unknown publish status response, abort.")
		}

	}

}

func (cm *cvmfsManager) UploadNewLayer(orig string) (layer ThinImageLayer, err error) {
	tarFileName, err := tar(orig)
	minioConfig = readConfig()

	if err != nil {
		fmt.Printf("Failed to create tar: %s\n", err.Error())
		return layer, err
	}

	h, err := sha256hash(tarFileName)
	if err != nil {
		fmt.Printf("Failed to calculate hash: %s\n", err.Error())
		return layer, err
	}

	fmt.Printf("Uploading file: %s\n", tarFileName)
	if err := upload(tarFileName, h); err != nil {
		fmt.Printf("Failed to upload: %s\n", err.Error())
		return layer, err
	}

	if err := os.Remove(tarFileName); err != nil {
		fmt.Printf("Couldn't remove the uploaded tmp tar.")
	}

	fmt.Println("Wait for it!")
	if err := waitForPublishing(h); err != nil {
		return layer, err
	}

	if err := cm.Remount(minioConfig.CvmfsRepo); err != nil {
		fmt.Printf("Failed to remount cvmfs repo: %s\n", err.Error())
		return layer, err
	}

	layer.Digest = h
	layer.Url = "cvmfs://" + minioConfig.CvmfsRepo + "/layers/" + h

	return layer, nil
}
