package aufs

import (
	"fmt"
	"github.com/docker/docker/pkg/parsers"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func (a *Driver) isThinImageLayer(id string) bool {
	diff_path := a.getDiffPath(id)
	magic_file_path := path.Join(diff_path, ".thin")
	_, err := os.Stat(magic_file_path)

	if err == nil {
		return true
	}

	return false
}

func (a *Driver) getCvmfsLayerPaths(id []string) []string {
	ret := make([]string, len(id))

	for i, layer := range id {
		ret[i] = path.Join(a.cvmfsMountPath, cvmfsDefaultRepo, "layers", layer)
	}

	return ret
}

func (a *Driver) appendCvmfsLayerPaths(oldArray []string, newArray []string) []string {
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

	fmt.Println(oldArray)
	fmt.Println(newArray)
	fmt.Println(ret)

	return ret
}

func (a *Driver) getNestedLayerIDs(id string) []string {
	magic_file_path := path.Join(a.getDiffPath(id), ".thin")
	content, _ := ioutil.ReadFile(magic_file_path)
	lines := strings.Split(string(content), "\n")
	return lines[:len(lines)-1]
}

func mountCvmfsRepo(repo, target string) error {
	mountTarget := path.Join(target, repo)
	os.MkdirAll(mountTarget, os.ModePerm)

	cmd := "cvmfs2 -o rw,fsname=cvmfs2,allow_other,grab_mountpoint"
	cmd += " " + repo
	cmd += " " + mountTarget

	// TODO: check for errors!
	exec.Command("bash", "-x", "-c", cmd).Run()

	return nil
}

func umountCvmfsRepo(target string) error {
	// TODO: check for errors!
	exec.Command("umount", target).Run()

	return nil
}

func (a *Driver) umountAllCvmfsRepos() error {
	umountCvmfsRepo(path.Join(a.cvmfsMountPath, cvmfsDefaultRepo))
	return nil
}

func (a *Driver) mountAllCvmfsRepos() error {
	mountCvmfsRepo(cvmfsDefaultRepo, a.cvmfsMountPath)
	return nil
}

func parseOptions(options []string) (map[string]string, error) {
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

func (a *Driver) configureCvmfs(options []string) error {
	m, err := parseOptions(options)

	if err != nil {
		return err
	}

	if method, ok := m["cvmfsMountMethod"]; !ok {
		a.cvmfsMountMethod = "internal"
	} else {
		a.cvmfsMountMethod = method
	}

	if mountPath, ok := m["cvmfsMountmethod"]; !ok {
		if a.cvmfsMountMethod == "internal" {
			a.cvmfsMountPath = "/cvmfs"
		} else if a.cvmfsMountMethod == "external" {
			a.cvmfsMountPath = "/mnt/cvmfs"
		}
	} else {
		a.cvmfsMountPath = mountPath
	}

	if a.cvmfsMountMethod == "external" &&
		!strings.HasPrefix(a.cvmfsMountPath, "/mnt") {
		return fmt.Errorf("CVMFS Mount path is not propagated!")
	}

	return nil
}
