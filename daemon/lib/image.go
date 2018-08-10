package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
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
		url = fmt.Sprintf("%s@%s", url, i.Digest)
	} else {
		url = fmt.Sprintf("%s%s", url, i.Tag)
	}
	return url
}

func (i Image) GetServerUrl() string {
	return fmt.Sprintf("%s://%s", i.Scheme, i.Registry)
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
	return manifest, nil
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
	fmt.Println("Auth to: " + WwwAuthenticate)
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
	Resp *http.Response
}

func (img Image) GetLayerIfNotInCVMFS(cvmfsRepo, subDir string, layers chan<- downloadedLayer) (err error) {
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
		// in this function before to wonwload something we check that the layer is not already in the repository
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
			layers <- toSend
		}
	}
	return nil

}

func (img Image) GetLayers(cvmfsRepo, subDir string, layers chan<- downloadedLayer) error {
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
		layers <- toSend
	}
	return nil
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
			Resp *http.Response
		}{layer.Digest, resp}
	} else {
		Log().Warning("Received status code ", resp.StatusCode)
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

	fmt.Println(req.URL)

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