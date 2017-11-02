package util

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/parsers"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

type ThinImageLayer struct {
	Digest   string `json:"digest"`
	Url      string `json:"url,omitempty"`
}

type ThinImage struct {
	Version    string           `json:"version"`
	MinVersion string           `json:"min_version,omitempty"`
	Origin     string           `json:"origin,omitempty"`
	Layers     []ThinImageLayer `json:"layers"`
	Comment    string           `json:"comment,omitempty"`
}

func (t *ThinImage) AddLayer(newLayer ThinImageLayer) {
	t.Layers = append([]ThinImageLayer{newLayer}, t.Layers...)
}

func reverse(in []ThinImageLayer) []ThinImageLayer {
	l := len(in)
	out := make([]ThinImageLayer, l)

	for i, v := range in {
		out[l-i-1] = v
	}

	return out
}

func IsThinImageLayer(diffPath string) bool {
	magic_file_path := path.Join(diffPath, "thin.json")
	_, err := os.Stat(magic_file_path)

	if err == nil {
		return true
	}

	return false
}


func ParseThinUrl(url string) (schema string, location string) {
	tokens := strings.Split(url, "://")
	if tokens[0] != "cvmfs" {
		fmt.Println("layer's url schema " + tokens[0] + " unsupported!" +
			          " [" + url + "]")
	}
	return tokens[0], strings.Join(tokens[1:], "://")
}


func ParseCvmfsLocation(location string) (repo string, folder string) {
	tokens := strings.Split(location, "/")
	return tokens[0], strings.Join(tokens[1:], "/")
}


func GetCvmfsLayerPaths(layers []ThinImageLayer, cvmfsMountPath string) []string {
	ret := make([]string, len(layers))

	for i, layer := range layers {
		root := cvmfsMountPath
		_, location := ParseThinUrl(layer.Url)
		repo, folder := ParseCvmfsLocation(location)

		ret[i] = path.Join(root, repo, folder)
	}

	return ret
}

func ExpandCvmfsLayerPaths(oldArray []string, newArray []string, i int) (result []string) {
	left := oldArray[:i]
	right := oldArray[i+1:]

	result = append(left, newArray...)
	result = append(result, right...)

	return result
}

func GetNestedLayerIDs(diffPath string) []ThinImageLayer {
	magic_file_path := path.Join(diffPath, "thin.json")
	content, _ := ioutil.ReadFile(magic_file_path)

	var thin ThinImage
	json.Unmarshal(content, &thin)

	return reverse(thin.Layers)
}

func ParseOptions(options []string) (map[string]string, error) {
	m := make(map[string]string)

	for _, v := range options {
		key, value, err := parsers.ParseKeyValueOpt(v)

		if err != nil {
			return nil, err
		}

		m[key] = value
	}

	return m, nil
}

func ReadThinFile(thinFilePath string) ThinImage {
	content, _ := ioutil.ReadFile(thinFilePath)
	var thin ThinImage
	json.Unmarshal(content, &thin)

	return thin
}

func WriteThinFile(thin ThinImage) (string, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	tmp := path.Join(os.TempDir(), fmt.Sprintf("dlcg-%d", rand.Int()))
	os.MkdirAll(tmp, os.ModePerm)

	p := path.Join(tmp, "thin.json")

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

type ICvmfsManager interface {
	GetLayers(layers ...ThinImageLayer) error
	PutLayers(layers ...ThinImageLayer) error
	PutAll() error
	Remount(repo string) error
	UploadNewLayer(string) (ThinImageLayer, error)
}

type cvmfsManager struct {
	mountPath string
	ctr       map[string]int
	mux       sync.Mutex
}

func NewCvmfsManager(cvmfsMountPath, cvmfsMountMethod string) ICvmfsManager {
	if cvmfsMountMethod == "external" {
		return nil
	}

	return &cvmfsManager{
		mountPath: cvmfsMountPath,
		ctr:       make(map[string]int),
	}
}

func (cm *cvmfsManager) mount(repo string) error {
	mountTarget := path.Join(cm.mountPath, repo)
	os.MkdirAll(mountTarget, os.ModePerm)

	cmd := "cvmfs2 -o rw,fsname=cvmfs2,allow_other,grab_mountpoint,cvmfs_suid"
	cmd += " '" + repo + "'"
	cmd += " '" + mountTarget + "'"

	// TODO: check for errors!
	out, err := exec.Command("bash", "-x", "-c", cmd).CombinedOutput()

	if err != nil {
		fmt.Println("There was an error during mount!")
		fmt.Printf("%s\n", out)
		return err
	} else {
		fmt.Println("Repo mounted successfully!")
	}

	return nil
}

func (cm *cvmfsManager) umount(repo string) error {
	// TODO: check for errors!
	fmt.Println("CernVM-FS: mounting %s", repo)
	mountTarget := path.Join(cm.mountPath, repo)
	cmd := exec.Command("umount", mountTarget)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Cvmfs mount failed.\nOutput:%s\nError%s\n",
			out, err)
	}

	return nil
}

func (cm *cvmfsManager) isConfigured(repo string) error {
	if strings.HasSuffix(repo, ".cern.ch") {
		return nil
	}

	fmt.Println("isConfigured()")

	confPath := "/etc/cvmfs/config.d"
	keysPath := "/etc/cvmfs/keys"

	repoConf := path.Join(confPath, repo) + ".conf"
	repoKeys := path.Join(keysPath, repo) + ".pub"

	errmsg1 := "Configuration for CVMFS repository %s is missing."
	errmsg2 := "Key for CVMFS repository %s is missing."

	if _, err := os.Stat(repoConf); os.IsNotExist(err) {
		return fmt.Errorf(errmsg1, repo)
	}

	if _, err := os.Stat(repoKeys); os.IsNotExist(err) {
		return fmt.Errorf(errmsg2, repo)
	}

	fmt.Println("Repo configuration ok!")

	return nil
}

func (cm *cvmfsManager) get(repo string) error {
	// TODO: maybe delegate this check to the mount call itself?
	if err := cm.isConfigured(repo); err != nil {
		return err
	}

	if _, ok := cm.ctr[repo]; !ok {
		fmt.Printf("Repo %s is not in the map!\n", repo)
		cm.ctr[repo] = 1
		cm.mount(repo)
		return nil
	} else {
		fmt.Printf("Repos %s is already in the map...\n", repo)
	}

	cm.ctr[repo] += 1

	return nil
}

func (cm *cvmfsManager) put(repo string) error {
	if _, ok := cm.ctr[repo]; !ok {
		return nil
	}

	if cm.ctr[repo] > 1 {
		cm.ctr[repo] -= 1
		return nil
	}

	cm.ctr[repo] = 0
	cm.umount(repo)
	return nil
}

func (cm *cvmfsManager) PutAll() error {
	cm.mux.Lock()
	defer cm.mux.Unlock()

	for k := range cm.ctr {
		cm.umount(k)
		cm.ctr[k] = 0
	}
	return nil
}

func (cm *cvmfsManager) GetLayers(layers ...ThinImageLayer) error {
	cm.mux.Lock()

	for i, l := range layers {
		_, location := ParseThinUrl(l.Url)
		repo, _ := ParseCvmfsLocation(location)
		if err := cm.get(repo); err != nil {
			cm.mux.Unlock()

			cm.PutLayers(layers[:i]...)
			return err
		}
	}

	cm.mux.Unlock()
	return nil
}

func (cm *cvmfsManager) PutLayers(layers ...ThinImageLayer) error {
	cm.mux.Lock()
	defer cm.mux.Unlock()

	for _, l := range layers {
		_, location := ParseThinUrl(l.Url)
		repo, _ := ParseCvmfsLocation(location)
		cm.put(repo)
	}
	return nil
}

func (cm *cvmfsManager) Remount(repo string) error {
	cm.mux.Lock()
	defer cm.mux.Unlock()

	if _, ok := cm.ctr[repo]; !ok {
		return nil
	}

	if cm.ctr[repo] == 0 {
		return nil
	}
	// TODO(jblomer): use net cat, cvmfs_talk unavailable in new image
	cmd := "cvmfs_talk"
	args := []string{"-i", repo, "remount", "sync"}

	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
		return err
	} else {
		return nil
	}
}
