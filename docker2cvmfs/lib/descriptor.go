package lib

import (
	"strings"

	cvmfsUtil "github.com/cvmfs/docker-graphdriver/plugins/util"
)

func MakeThinImage(m Manifest, repo string) cvmfsUtil.ThinImage {
	layers := make([]cvmfsUtil.ThinImageLayer, len(m.Layers))

	for i, l := range m.Layers {
		d := strings.Split(l.Digest, ":")[1]
		layers[i] = cvmfsUtil.ThinImageLayer{Digest: d, Repo: repo}
	}

	return cvmfsUtil.ThinImage{Layers: layers, Version: thinImageVersion}
}
