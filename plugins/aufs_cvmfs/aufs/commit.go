package aufs

import (
	"errors"
	"path"

	"github.com/cvmfs/docker-graphdriver/plugins/util"
)

func (a *Driver) getParentThinLayer(id string) (util.ThinImage, error) {
	roLayers, _ := getParentIDs(a.rootPath(), id)
	var thin util.ThinImage

	for _, l := range roLayers {
		diffPath := a.getDiffPath(l)
		if util.IsThinImageLayer(diffPath) {
			return util.ReadThinFile(path.Join(diffPath, ".thin.json")), nil
		}
	}

	return thin, errors.New("Not a thin image!")
}
