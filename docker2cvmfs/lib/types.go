package lib

type TokenMessage struct {
	Token string
}

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
