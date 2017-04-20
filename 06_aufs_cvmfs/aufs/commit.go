package aufs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/atlantic777/docker_graphdriver_plugins/util"
)

func (a *Driver) getParentThinLayer(id string) (util.ThinImage, error) {
	roLayers, _ := getParentIDs(a.rootPath(), id)
	var thin util.ThinImage

	for _, l := range roLayers {
		diffPath := a.getDiffPath(l)
		if util.IsThinImageLayer(l, diffPath) {
			return a.readThinFile(l), nil
		}
	}

	return thin, errors.New("Not a thin image!")
}

func (a *Driver) readThinFile(id string) util.ThinImage {
	thin_file_path := path.Join(a.getDiffPath(id), ".thin")
	content, _ := ioutil.ReadFile(thin_file_path)

	var thin util.ThinImage
	json.Unmarshal(content, &thin)

	return thin
}

func (a *Driver) writeThinFile(thin util.ThinImage, id string) (string, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	tmp := path.Join(os.TempDir(), fmt.Sprintf("dlcg-%d", rand.Int()))
	os.MkdirAll(tmp, os.ModePerm)

	p := path.Join(tmp, ".thin")

	j, err := json.Marshal(thin)
	if err != nil {
		fmt.Println("Failed to marshal new thin file to json!")
		return "", err
	}

	if err := ioutil.WriteFile(p, j, os.ModePerm); err != nil {
		fmt.Printf("Failed to write new thin file!\nFile: %s\n", p)
		return "", err
	}

	return tmp, nil
}
