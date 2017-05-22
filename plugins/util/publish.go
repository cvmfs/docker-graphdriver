package util

import (
	"bytes"
	"crypto/sha256"
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

var (
	accessKey = "LD9GJ1V2T9V6LA09E052"
	secretKey = "CgrYAbowR91YJykDqvWX3kxBQIzvGKZ20n2l7MwY"
	host      = "nhardi-stratum0.cern.ch:9000"
	ssl       = false
)

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
	minioClient, err := minio.New(host, accessKey, secretKey, ssl)
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

func remountCvmfs() error {
	return exec.Command("cvmfs_talk", "remount", "sync").Run()
}

func UploadNewLayer(orig string) (layer ThinImageLayer, err error) {
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

	if err := remountCvmfs(); err != nil {
		fmt.Printf("Failed to remount cvmfs repo: %s\n", err.Error())
		return layer, err
	}

	layer.Digest = h
	layer.Repo = "nhardi-ansible.cern.ch"

	return layer, nil
}
