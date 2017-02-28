package aufs

import (
	"fmt"
	"io/ioutil"
	"os"
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
		ret[i] = path.Join(a.cvmfsRoot, layer)
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
