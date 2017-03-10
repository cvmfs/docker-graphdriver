package aufs

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
	// tmp, err := ioutil.TempDir(os.TempDir(), "dlcg-")
	// if err != nil {
	// 	fmt.Println("Failed to create temp dir.")
	// 	return "", err
	// }
	rand.Seed(time.Now().UTC().UnixNano())
	tmp := path.Join(os.TempDir(), fmt.Sprintf("dlcg-%d", rand.Int()))

	// if err := os.Rename(src, tmp); err != nil {
	// 	fmt.Println("Failed to rename")
	// 	return "", err
	// }
	var out bytes.Buffer
	cmd := fmt.Sprintf("mv %s/* %s", src, tmp)
	c := exec.Command("bash", "-c", cmd)
	c.Stdout = &out
	if err := c.Run(); err != nil {
		fmt.Println("Failed to move!")
		return "", err
	}

	if err := os.Mkdir(src, os.ModePerm); err != nil {
		fmt.Println("Failed to recreate original dir.")
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

func hash(src string) (string, error) {
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

func upload(src, hash string) error {
	minioClient, err := minio.New(host, accessKey, secretKey, ssl)
	if err != nil {
		fmt.Println("Failed to create a minio client!")
		return err
	}

	_, err = minioClient.FPutObject("layers", hash, src, "application/x-gzip")
	if err != nil {
		fmt.Println("Failed call to FPutObject()")
		return err
	}

	return nil
}

func MoveAndUpload(orig string) (string, error) {
	tmpDirectory, err := move(orig)
	if err != nil {
		fmt.Printf("Failed to move directory: %s\n", err.Error())
		return "", err
	}

	tarFileName, err := tar(tmpDirectory)
	if err != nil {
		fmt.Printf("Failed to create tar: %s\n", err.Error())
		return "", err
	}

	h, err := hash(tarFileName)
	if err != nil {
		fmt.Printf("Failed to calculate hash: %s\n", err.Error())
		return "", err
	}

	if err := upload(tarFileName, h); err != nil {
		fmt.Printf("Failed to upload: %s\n", err.Error())
		return "", err
	}

	return h, nil
}
