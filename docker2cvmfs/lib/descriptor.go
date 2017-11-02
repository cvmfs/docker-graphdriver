package lib

import (
	"strings"

	cvmfsUtil "github.com/cvmfs/docker-graphdriver/plugins/util"
)

func MakeThinImage(m Manifest, repoLocation string, origin string) cvmfsUtil.ThinImage {
	layers := make([]cvmfsUtil.ThinImageLayer, len(m.Layers))

	url_base := "cvmfs://" + repoLocation
	for i, l := range m.Layers {
		d := strings.Split(l.Digest, ":")[1]
		url := url_base + "/" + d;
		layers[i] = cvmfsUtil.ThinImageLayer{Digest: d, Url: url}
	}

	return cvmfsUtil.ThinImage{Layers: layers,
		                         Origin: origin,
		                         Version: thinImageVersion}
}
