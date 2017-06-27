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
	"os"
	"os/exec"
	"path"
	"time"
)

type MinioConfig struct {
	AccessKey    string
	AccessSecret string
	Host         string
	CvmfsRepo    string
	SSL          bool
}

var minioConfig = readConfig()

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

	_, err = minioClient.FPutObject("layers", h, src, "application/x-gzip")
	if err != nil {
		fmt.Println("Failed call to FPutObject()")
		return err
	}

	return nil
}

func (cm *cvmfsManager) UploadNewLayer(orig string) (layer ThinImageLayer, err error) {
	tarFileName, err := tar(orig)

	if err != nil {
		fmt.Printf("Failed to create tar: %s\n", err.Error())
		return layer, err
	}

	h, err := sha256hash(tarFileName)
	if err != nil {
		fmt.Printf("Failed to calculate hash: %s\n", err.Error())
		return layer, err
	}

	if err := upload(tarFileName, h); err != nil {
		fmt.Printf("Failed to upload: %s\n", err.Error())
		return layer, err
	}

	if err := cm.Remount(minioConfig.CvmfsRepo); err != nil {
		fmt.Printf("Failed to remount cvmfs repo: %s\n", err.Error())
		return layer, err
	}

	layer.Digest = h
	layer.Repo = minioConfig.CvmfsRepo

	return layer, nil
}
