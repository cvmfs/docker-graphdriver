package aufs

import (
	"errors"
	"path"

	"github.com/atlantic777/docker_graphdriver_plugins/util"
)

func (a *Driver) getParentThinLayer(id string) (util.ThinImage, error) {
	roLayers, _ := getParentIDs(a.rootPath(), id)
	var thin util.ThinImage

	for _, l := range roLayers {
		diffPath := a.getDiffPath(l)
		if util.IsThinImageLayer(diffPath) {
			return util.ReadThinFile(path.Join(diffPath, ".thin")), nil
		}
	}

	return thin, errors.New("Not a thin image!")
}
