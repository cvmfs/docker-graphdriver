package lib

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/docker/docker/image"
	"github.com/olekukonko/tablewriter"
	copy "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	d2c "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
)

type ManifestRequest struct {
	Image    Image
	Password string
}

type Image struct {
	Id         int
	User       string
	Scheme     string
	Registry   string
	Repository string
	Tag        string
	Digest     string
	IsThin     bool
	Manifest   *d2c.Manifest
}

func (i Image) GetSimpleName() string {
	name := fmt.Sprintf("%s/%s", i.Registry, i.Repository)
	if i.Tag == "" {
		return name
	} else {
		return name + ":" + i.Tag
	}
}

func (i Image) WholeName() string {
	root := fmt.Sprintf("%s://%s/%s", i.Scheme, i.Registry, i.Repository)
	if i.Tag != "" {
		root = fmt.Sprintf("%s:%s", root, i.Tag)
	}
	if i.Digest != "" {
		root = fmt.Sprintf("%s@%s", root, i.Digest)
	}
	return root
}

func (i Image) GetManifestUrl() string {
	url := fmt.Sprintf("%s://%s/v2/%s/manifests/", i.Scheme, i.Registry, i.Repository)
	if i.Digest != "" {
		url = fmt.Sprintf("%s%s", url, i.Digest)
	} else {
		url = fmt.Sprintf("%s%s", url, i.Tag)
	}
	return url
}

func (i Image) GetServerUrl() string {
	return fmt.Sprintf("%s://%s", i.Scheme, i.Registry)
}

func (i Image) GetReference() string {
	if i.Digest == "" && i.Tag != "" {
		return ":" + i.Tag
	}
	if i.Digest != "" && i.Tag == "" {
		return "@" + i.Digest
	}
	if i.Digest != "" && i.Tag != "" {
		return ":" + i.Tag + "@" + i.Digest
	}
	panic("Image wrong format, missing both tag and digest")
}

func (i Image) GetSimpleReference() string {
	if i.Tag != "" {
		return i.Tag
	}
	if i.Digest != "" {
		return i.Digest
	}
	panic("Image wrong format, missing both tag and digest")
}

func (img Image) PrintImage(machineFriendly, csv_header bool) {
	if machineFriendly {
		if csv_header {
			fmt.Printf("name,user,scheme,registry,repository,tag,digest,is_thin\n")
		}
		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			img.WholeName(), img.User, img.Scheme,
			img.Registry, img.Repository,
			img.Tag, img.Digest,
			fmt.Sprint(img.IsThin))
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Key", "Value"})
		table.Append([]string{"Name", img.WholeName()})
		table.Append([]string{"User", img.User})
		table.Append([]string{"Scheme", img.Scheme})
		table.Append([]string{"Registry", img.Registry})
		table.Append([]string{"Repository", img.Repository})
		table.Append([]string{"Tag", img.Tag})
		table.Append([]string{"Digest", img.Digest})
		var is_thin string
		if img.IsThin {
			is_thin = "true"
		} else {
			is_thin = "false"
		}
		table.Append([]string{"IsThin", is_thin})
		table.Render()
	}
}

func (img Image) GetManifest() (d2c.Manifest, error) {
	if img.Manifest != nil {
		return *img.Manifest, nil
	}
	bytes, err := img.getByteManifest()
	if err != nil {
		return d2c.Manifest{}, err
	}
	var manifest d2c.Manifest
	err = json.Unmarshal(bytes, &manifest)
	if err != nil {
		return manifest, err
	}
	if reflect.DeepEqual(d2c.Manifest{}, manifest) {
		return manifest, fmt.Errorf("Got empty manifest")
	}
	img.Manifest = &manifest
	return manifest, nil
}

func (img Image) GetChanges() (changes []string, err error) {
	user := img.User
	pass, err := GetPassword(img.User, img.Registry)
	if err != nil {
		LogE(err).Warning("Unable to get the credential for downloading the configuration blog, trying anonymously")
		user = ""
		pass = ""
	}

	changes = []string{"ENV CVMFS_IMAGE true"}
	manifest, err := img.GetManifest()
	if err != nil {
		LogE(err).Warning("Impossible to retrieve the manifest of the image, not changes set")
		return
	}
	configUrl := fmt.Sprintf("%s://%s/v2/%s/blobs/%s",
		img.Scheme, img.Registry, img.Repository, manifest.Config.Digest)
	token, err := firstRequestForAuth(configUrl, user, pass)
	if err != nil {
		LogE(err).Warning("Impossible to retrieve the token for getting the changes from the repository, not changes set")
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", configUrl, nil)
	if err != nil {
		LogE(err).Warning("Impossible to create a request for getting the changes no chnages set.")
		return
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogE(err).Warning("Error in reading the body from the configuration, no change set")
		return
	}

	var config image.Image
	err = json.Unmarshal(body, &config)
	if err != nil {
		LogE(err).Warning("Error in unmarshaling the configuration of the image")
		return
	}
	env := config.Config.Env

	if len(env) > 0 {
		for _, e := range env {
			envs := strings.SplitN(e, "=", 2)
			if len(envs) != 2 {
				continue
			}
			change := fmt.Sprintf("ENV %s=\"%s\"", envs[0], envs[1])
			changes = append(changes, change)
		}
	}

	cmd := config.Config.Cmd

	if len(cmd) > 0 {
		for _, c := range cmd {
			changes = append(changes, fmt.Sprintf("CMD %s", c))
		}
	}

	return
}

func (img Image) GetSingularityLocation() string {
	return fmt.Sprintf("docker://%s/%s%s", img.Registry, img.Repository, img.GetReference())
}

type Singularity struct {
	Image         *Image
	TempDirectory string
}

func (img Image) DownloadSingularityDirectory() (sing Singularity, err error) {
	dir, err := ioutil.TempDir("", "singularity_buffer")
	if err != nil {
		return
	}
	err = ExecCommand("singularity", "build", "--sandbox", dir, img.GetSingularityLocation())
	if err != nil {
		LogE(err).Error("Error in downloading the singularity image")
		return
	}

	Log().Info("Successfully download the singularity image")
	return Singularity{Image: &img, TempDirectory: dir}, nil
}

func (s Singularity) IngestIntoCVMFS(CVMFSRepo string) error {
	defer func() {
		Log().WithFields(log.Fields{"image": s.Image.GetSimpleName(), "action": "ingesting singularity"}).Info("Deleting temporary direcotry")
		os.RemoveAll(s.TempDirectory)
	}()
	Log().WithFields(log.Fields{"image": s.Image.GetSimpleName(), "action": "ingesting singularity"}).Info("Start ingesting")
	path := filepath.Join("/", "cvmfs", CVMFSRepo, ".singularity", s.Image.Registry, s.Image.Repository, s.Image.GetSimpleReference())

	Log().WithFields(log.Fields{"image": s.Image.GetSimpleName(), "action": "ingesting singularity"}).Info("Start ingesting")
	err := ExecCommand("cvmfs_server", "transaction", CVMFSRepo)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Error("Error in opening the transaction")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}

	Log().WithFields(log.Fields{"image": s.Image.GetSimpleName(), "action": "ingesting singularity"}).Info("Copying directory")
	os.RemoveAll(path)
	err = os.MkdirAll(path, 0666)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Warning("Error in creating the directory where to copy the singularity")
	}
	err = copy.Copy(s.TempDirectory, path)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Error("Error in moving the directory inside the CVMFS repo")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}

	Log().WithFields(log.Fields{"image": s.Image.GetSimpleName(), "action": "ingesting singularity"}).Info("Publishing")
	err = ExecCommand("cvmfs_server", "publish", CVMFSRepo)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Error("Error in publishing the repository")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}

	return nil
}

func (img Image) getByteManifest() ([]byte, error) {
	pass, err := GetPassword(img.User, img.Registry)
	if err != nil {
		LogE(err).Warning("Unable to retrieve the password, trying to get the manifest anonymously.")
		return img.getAnonymousManifest()
	}
	return img.getManifestWithPassword(pass)
}

func (img Image) getAnonymousManifest() ([]byte, error) {
	return getManifestWithUsernameAndPassword(img, "", "")
}

func (img Image) getManifestWithPassword(password string) ([]byte, error) {
	return getManifestWithUsernameAndPassword(img, img.User, password)
}

func getManifestWithUsernameAndPassword(img Image, user, pass string) ([]byte, error) {

	url := img.GetManifestUrl()

	token, err := firstRequestForAuth(url, user, pass)
	if err != nil {
		LogE(err).Error("Error in getting the authentication token")
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		LogE(err).Error("Impossible to create a HTTP request")
		return nil, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := client.Do(req)
	if err != nil {
		LogE(err).Error("Error in making the HTTP request")
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogE(err).Error("Error in reading the second http response")
		return nil, err
	}
	return body, nil
}

func firstRequestForAuth(url, user, pass string) (token string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		LogE(err).Error("Error in making the first request for auth")
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 401 {
		log.WithFields(log.Fields{
			"status code": resp.StatusCode,
		}).Info("Expected status code 401, print body anyway.")
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			LogE(err).Error("Error in reading the first http response")
		}
		fmt.Println(string(body))
		return "", err
	}
	WwwAuthenticate := resp.Header["Www-Authenticate"][0]
	token, err = requestAuthToken(WwwAuthenticate, user, pass)
	if err != nil {
		LogE(err).Error("Error in getting the authentication token")
		return "", err
	}
	return token, nil

}

func getLayerUrl(img Image, layer d2c.Layer) string {
	return fmt.Sprintf("%s://%s/v2/%s/blobs/%s",
		img.Scheme, img.Registry, img.Repository, layer.Digest)
}

type downloadedLayer struct {
	Name string
	Resp io.ReadCloser
}

func (img Image) GetLayerIfNotInCVMFS(cvmfsRepo, subDir string, layers chan<- downloadedLayer, stopGettingLayers <-chan bool) (err error) {
	// remember to close the layers channel when done
	defer close(layers)
	// get the credential
	user := img.User
	pass, err := GetPassword(img.User, img.Registry)
	if err != nil {
		LogE(err).Warning("Unable to retrieve the password, trying to get the layers anonymously.")
		user = ""
		pass = ""
	}

	// then we try to get the manifest from our database
	manifest, err := img.GetManifest()
	if err != nil {
		LogE(err).Warn("Error in getting the manifest")
		return err
	}

	// A first request is used to get the authentication
	firstLayer := manifest.Layers[0]
	layerUrl := getLayerUrl(img, firstLayer)
	token, err := firstRequestForAuth(layerUrl, user, pass)
	if err != nil {
		return err
	}
	// at this point we iterate each layer and we download it.
	for _, layer := range manifest.Layers {
		// in this function before to donwload something we check that the layer is not already in the repository
		layerDigest := strings.Split(layer.Digest, ":")[1]
		location := filepath.Join("/", "cvmfs", cvmfsRepo, subDir, layerDigest)
		_, err := os.Stat(location)
		if err == nil {
			// the path exists
			Log().WithFields(log.Fields{"layer": layer.Digest}).Info("Layer already exists, skipping download")
			continue
		}
		toSend, err := img.downloadLayer(layer, token)
		if err != nil {
			LogE(err).Error("Error in downloading a layer")
		} else {
			select {
			case layers <- toSend:
				{
				}
			case <-stopGettingLayers:
				Log().Info("Receive stop signal")
				return nil
			}
		}
	}
	return nil

}

func (img Image) GetLayers(cvmfsRepo, subDir string, layers chan<- downloadedLayer, stopGettingLayers <-chan bool) error {
	// first we get the username and password, if we are not able to get those,
	// we try anyway anonymously
	defer close(layers)
	user := img.User
	pass, err := GetPassword(img.User, img.Registry)
	if err != nil {
		LogE(err).Warning("Unable to retrieve the password, trying to get the layers anonymously.")
		user = ""
		pass = ""
	}

	// then we try to get the manifest from our database
	manifest, err := img.GetManifest()
	if err != nil {
		LogE(err).Warn("Error in getting the manifest")
		return err
	}

	// A first request is used to get the authentication
	firstLayer := manifest.Layers[0]
	layerUrl := getLayerUrl(img, firstLayer)
	token, err := firstRequestForAuth(layerUrl, user, pass)
	if err != nil {
		return err
	}
	// at this point we iterate each layer and we download it.
	for _, layer := range manifest.Layers {
		toSend, err := img.downloadLayer(layer, token)
		if err != nil {
			LogE(err).Error("Error in downloading a layer")
		}
		select {
		case layers <- toSend:
			{
			}
		case <-stopGettingLayers:
			Log().Info("Received stop signal")
			return nil
		}
	}
	return nil
}

func (img Image) downloadImage() (err error) {
	err = ExecCommand("docker", "pull", img.GetSimpleName())
	if err != nil {
		LogE(err).Error("Error in pulling from the registry")
		return
	}
	return
}

func (img Image) saveDockerLayerAndManifestOnDisk(dockerSavedImage string) (manifest string, paths []string, err error) {

	f, err := ioutil.TempFile("", "dockerSave")
	defer f.Close()
	if err != nil {
		LogE(err).Error("Error in creating the temp file where to save the image")
		return
	}
	err = ExecCommand("docker", "save", img.GetSimpleName(), "--output", f.Name())
	if err != nil {
		defer os.Remove(f.Name())
		LogE(err).Error("Error in saving the image")
		return
	}

	deleteFiles := func() {
		if manifest != "" {
			os.Remove(manifest)
		}
		if len(paths) != 0 {
			for _, p := range paths {
				os.Remove(p)
			}
		}
	}
	tempDir, err := ioutil.TempDir("", "layers")
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			LogE(err).Error("Error in reading the archive")
			deleteFiles()
			return "", nil, err
		}
		if hdr.Name == "manifest.json" {
			tempManifest, err := ioutil.TempFile(tempDir, "tempManifest")
			if err != nil {
				LogE(err).Error("Error in creating manifest temp file")
				deleteFiles()
				return "", nil, err
			}
			manifest = tempManifest.Name()
			_, err = io.Copy(f, tr)
			if err != nil {
				LogE(err).Error("Error in copying the manifest to a file")
				deleteFiles()
				return "", nil, err
			}
		}
		// the files are stored inside the directory $layer (so the actual name of the layer) as tar archive called "layer.tar", not the layer name, the string "layer"
		if strings.HasSuffix(hdr.Name, ".tar") {
			splitList := filepath.SplitList(hdr.Name)
			splitListLen := len(splitList)
			if splitListLen < 2 {
				err = fmt.Errorf("Error in reading the tar for layer, less than 2 subpath for layer, expected at least two: $layerName/layer.tar")
				LogE(err).WithFields(log.Fields{"name": hdr.Name}).Error(err)
				deleteFiles()
				return "", nil, err
			}
			layerName := splitList[splitListLen-2]
			fLayer, err := os.OpenFile(filepath.Join(tempDir, layerName, ".tar"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
			if err != nil {
				LogE(err).Error("Impossible to exclusively open the file to write the tar layer")
				deleteFiles()
				return "", nil, err
			}
			_, err = io.Copy(fLayer, tr)
			if err != nil {
				LogE(err).Error("Error in copying the layer in a file")
				deleteFiles()
				return "", nil, err
			}
			paths = append(paths, f.Name())
			f.Close()
		}
	}
	return
}

func (img Image) removeImage() (err error) {
	err = ExecCommand("docker", "rmi", img.GetSimpleName())
	if err != nil {
		LogE(err).Error("Error in removing the image")
		return
	}
	return
}

func (img Image) downloadLayer(layer d2c.Layer, token string) (toSend downloadedLayer, err error) {
	user := img.User
	pass, err := GetPassword(img.User, img.Registry)
	if err != nil {
		LogE(err).Warning("Unable to retrieve the password, trying to get the layers anonymously.")
		user = ""
		pass = ""
	}
	layerUrl := getLayerUrl(img, layer)
	if token == "" {
		token, err = firstRequestForAuth(layerUrl, user, pass)
		if err != nil {
			return
		}
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", layerUrl, nil)
	if err != nil {
		LogE(err).Error("Impossible to create the HTTP request.")
		return
	}
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)
	Log().WithFields(log.Fields{"layer": layer.Digest}).Info("Make request for layer")
	if err != nil {
		return
	}
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		toSend = struct {
			Name string
			Resp io.ReadCloser
		}{layer.Digest, resp.Body}
	} else {
		Log().Warning("Received status code ", resp.StatusCode)
		// TODO add error
		err = fmt.Errorf("Layer not received, status code: %s", resp.StatusCode)
	}
	return

}

func parseBearerToken(token string) (realm string, options map[string]string, err error) {
	options = make(map[string]string)
	args := token[7:]
	keyValue := strings.Split(args, ",")
	for _, kv := range keyValue {
		splitted := strings.Split(kv, "=")
		if len(splitted) != 2 {
			err = fmt.Errorf("Wrong formatting of the token")
			return
		}
		splitted[1] = strings.Trim(splitted[1], `"`)
		if splitted[0] == "realm" {
			realm = splitted[1]
		} else {
			options[splitted[0]] = splitted[1]
		}
	}
	return
}

func requestAuthToken(token, user, pass string) (authToken string, err error) {
	realm, options, err := parseBearerToken(token)
	if err != nil {
		return
	}
	req, err := http.NewRequest("GET", realm, nil)
	if err != nil {
		return
	}

	query := req.URL.Query()
	for k, v := range options {
		query.Add(k, v)
	}
	if user != "" && pass != "" {
		query.Add("offline_token", "true")
		req.SetBasicAuth(user, pass)
	}
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = fmt.Errorf("Authorization error %s", resp.Status)
		return
	}

	var jsonResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	authTokenInterface, ok := jsonResp["token"]
	if ok {
		authToken = "Bearer " + authTokenInterface.(string)
	} else {
		err = fmt.Errorf("Didn't get the token key from the server")
		return
	}
	return
}
