package dockerutil

import (
	"strings"
)

type ConfigType struct {
	MediaType string
	Size      int
	Digest    string
}

type Layer struct {
	MediaType string
	Size      int
	Digest    string
}

type Manifest struct {
	SchemaVersion int
	MediaType     string
	Config        ConfigType
	Layers        []Layer
}

type ThinImageLayer struct {
	Digest string `json:"digest"`
	Url    string `json:"url,omitempty"`
}

type ThinImage struct {
	Version    string           `json:"version"`
	MinVersion string           `json:"min_version,omitempty"`
	Origin     string           `json:"origin,omitempty"`
	Layers     []ThinImageLayer `json:"layers"`
	Comment    string           `json:"comment,omitempty"`
}

var thinImageVersion = "1.0"

// m is the manifest of the original image
// repoLocation is where inside the repo we saved the several layers
// origin is an ecoding fo the original referencese and original registry
// I believe origin is quite useless but maybe is better to preserv it for
// ergonomic reasons.

func MakeThinImage(m Manifest, repoLocation string, origin string) ThinImage {
	layers := make([]ThinImageLayer, len(m.Layers))

	url_base := "cvmfs://" + repoLocation
	for i, l := range m.Layers {
		d := strings.Split(l.Digest, ":")[1]
		url := url_base + "/" + d
		layers[i] = ThinImageLayer{Digest: d, Url: url}
	}

	return ThinImage{Layers: layers,
		Origin:  origin,
		Version: thinImageVersion}
}
