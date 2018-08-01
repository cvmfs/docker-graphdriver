package lib

import (
	"strings"

	cvmfsUtil "github.com/cvmfs/docker-graphdriver/plugins/util"
)

// m is the manifest of the original image
// repoLocation is where inside the repo we saved the several layers
// origin is an ecoding fo the original referencese and original registry
// I believe origin is quite useless but maybe is better to preserv it for
// ergonomic reasons.
func MakeThinImage(m Manifest, repoLocation string, origin string) cvmfsUtil.ThinImage {
	layers := make([]cvmfsUtil.ThinImageLayer, len(m.Layers))

	url_base := "cvmfs://" + repoLocation
	for i, l := range m.Layers {
		d := strings.Split(l.Digest, ":")[1]
		url := url_base + "/" + d
		layers[i] = cvmfsUtil.ThinImageLayer{Digest: d, Url: url}
	}

	return cvmfsUtil.ThinImage{Layers: layers,
		Origin:  origin,
		Version: thinImageVersion}
}
