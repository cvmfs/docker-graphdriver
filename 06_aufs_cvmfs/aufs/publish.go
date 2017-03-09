package aufs

import (
	"crypto/sha256"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"github.com/minio/minio-go"
	"io"
	"io/ioutil"
	"os"
)

var (
	accessKey = "LD9GJ1V2T9V6LA09E052"
	secretKey = "CgrYAbowR91YJykDqvWX3kxBQIzvGKZ20n2l7MwY"
	host      = "nhardi-stratum0.cern.ch:9000"
	ssl       = false
)

func move(src string) (dst string) {
	tmp, _ := ioutil.TempDir(os.TempDir(), "dlcg-")

	os.Rename(src, tmp)
	os.Mkdir(src, os.ModePerm)

	return tmp
}

func tar(src string) string {
	dstFile, _ := ioutil.TempFile(os.TempDir(), "dlcg-tar-")
	defer dstFile.Close()

	tarReader, err := archive.Tar(src, archive.Gzip)

	if err != nil {
		panic(err)
	}

	io.Copy(dstFile, tarReader)

	return dstFile.Name()
}

func hash(src string) string {
	tarFile, _ := os.Open(src)
	defer tarFile.Close()

	h := sha256.New()
	io.Copy(h, tarFile)
	s := fmt.Sprintf("%x", h.Sum(nil))

	return s
}

func upload(src, hash string) {
	minioClient, err := minio.New(host, accessKey, secretKey, ssl)

	if err != nil {
		panic(err)
	}

	n, err := minioClient.FPutObject("layers", hash, src, "application/x-gzip")
	fmt.Println(n)

	if err != nil {
		panic(err)
	}

}

func MoveAndUpload(orig string) string {
	tmpDirectory := move(orig)
	tarFileName := tar(tmpDirectory)
	h := hash(tarFileName)
	upload(tarFileName, h)

	return h
}
