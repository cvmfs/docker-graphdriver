package util

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/parsers"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

var (
	cvmfsDefaultRepo = "docker2cvmfs-ci.cern.ch"
)

type ThinImageLayer struct {
	Digest   string `json:"digest"`
	Repo     string `json:"repo,omitempty"`
	Location string `json:"location,omitempty"`
}

type ThinImage struct {
	Version    string           `json:"version"`
	MinVersion string           `json:"min_version,omitempty"`
	Layers     []ThinImageLayer `json:"layers"`
	Comment    string           `json:"comment,omitempty"`
}

func (t *ThinImage) AddLayer(id string) {
	newLayer := ThinImageLayer{Digest: id}
	t.Layers = append([]ThinImageLayer{newLayer}, t.Layers...)
}

func IsThinImageLayer(diffPath string) bool {
	magic_file_path := path.Join(diffPath, ".thin")
	_, err := os.Stat(magic_file_path)

	if err == nil {
		return true
	}

	return false
}

func GetCvmfsLayerPaths(layers []ThinImageLayer, cvmfsMountPath string) []string {
	ret := make([]string, len(layers))

	for i, layer := range layers {
		root := cvmfsMountPath
		repo := cvmfsDefaultRepo
		location := "layers"
		digest := layer.Digest

		ret[i] = path.Join(root, repo, location, digest)
	}

	return ret
}

func AppendCvmfsLayerPaths(oldArray []string, newArray []string) []string {
	l_old := len(oldArray)
	l_new := len(newArray)
	newLen := l_old + l_new - 1
	ret := make([]string, newLen)

	for i := 0; i < l_old-1; i++ {
		ret[i] = oldArray[i]
	}

	for i := range newArray {
		ret[l_old-1+i] = newArray[i]
	}

	return ret
}

func GetNestedLayerIDs(diffPath string) []ThinImageLayer {
	magic_file_path := path.Join(diffPath, ".thin")
	content, _ := ioutil.ReadFile(magic_file_path)

	var thin ThinImage
	json.Unmarshal(content, &thin)

	return thin.Layers
}

func mountCvmfsRepo(repo, target string) error {
	mountTarget := path.Join(target, repo)
	os.MkdirAll(mountTarget, os.ModePerm)

	cmd := "cvmfs2 -o rw,fsname=cvmfs2,allow_other,grab_mountpoint"
	cmd += " " + repo
	cmd += " " + mountTarget

	// TODO: check for errors!
	out, err := exec.Command("bash", "-x", "-c", cmd).Output()

	if err != nil {
		fmt.Println("There was an error during mount!")
		fmt.Printf("%s\n", out)
		return err
	} else {
		fmt.Println("default repo mounted successfully!")
	}

	return nil
}

func umountCvmfsRepo(target string) error {
	// TODO: check for errors!
	exec.Command("umount", target).Run()

	return nil
}

func UmountAllCvmfsRepos(cvmfsMountPath string) error {
	umountCvmfsRepo(path.Join(cvmfsMountPath, cvmfsDefaultRepo))
	return nil
}

func MountAllCvmfsRepos(cvmfsMountPath string) error {
	mountCvmfsRepo(cvmfsDefaultRepo, cvmfsMountPath)
	return nil
}

func ParseOptions(options []string) (map[string]string, error) {
	m := make(map[string]string)

	for _, v := range options {
		key, value, err := parsers.ParseKeyValueOpt(v)

		if err != nil {
			return nil, err
		}

		m[key] = value
	}

	return m, nil
}
