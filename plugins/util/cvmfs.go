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
	"time"
)

type ThinImageLayer struct {
	Digest   string `json:"digest"`
	Repo     string `json:"repo,omitempty"`
	Location string `json:"location,omitempty"`
}

type ThinImage struct {
	Version    string           `json:"version"`
	MinVersion string           `json:"min_version,omitempty"`
	Layers     []ThinImageLayer `json:"layers"`
	Comment    string           `json:"comment,omitempty"`
}

func (t *ThinImage) AddLayer(newLayer ThinImageLayer) {
	t.Layers = append([]ThinImageLayer{newLayer}, t.Layers...)
}

func IsThinImageLayer(diffPath string) bool {
	magic_file_path := path.Join(diffPath, ".thin")
	_, err := os.Stat(magic_file_path)

	if err == nil {
		return true
	}

	return false
}

func GetCvmfsLayerPaths(layers []ThinImageLayer, cvmfsMountPath string) []string {
	ret := make([]string, len(layers))

	for i, layer := range layers {
		root := cvmfsMountPath
		// repo := cvmfsDefaultRepo
		repo := layer.Repo
		location := "layers"
		digest := layer.Digest

		ret[i] = path.Join(root, repo, location, digest)
	}

	return ret
}

func AppendCvmfsLayerPaths(oldArray []string, newArray []string) []string {
	l_old := len(oldArray)
	l_new := len(newArray)
	newLen := l_old + l_new - 1
	ret := make([]string, newLen)

	for i := 0; i < l_old-1; i++ {
		ret[i] = oldArray[i]
	}

	for i := range newArray {
		ret[l_old-1+i] = newArray[i]
	}

	return ret
}

func GetNestedLayerIDs(diffPath string) []ThinImageLayer {
	magic_file_path := path.Join(diffPath, ".thin")
	content, _ := ioutil.ReadFile(magic_file_path)

	var thin ThinImage
	json.Unmarshal(content, &thin)

	return thin.Layers
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

	p := path.Join(tmp, ".thin")

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
	Get(repo string) error
	GetLayers(layers ...ThinImageLayer) error
	Put(repo string) error
	PutLayers(layers ...ThinImageLayer) error
	PutAll() error
	Remount(repo string) error
}

type cvmfsManager struct {
	mountPath string
	ctr       map[string]int
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

	cmd := "cvmfs2 -o rw,fsname=cvmfs2,allow_other,grab_mountpoint"
	cmd += " " + repo
	cmd += " " + mountTarget

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
	mountTarget := path.Join(cm.mountPath, repo)
	cmd := exec.Command("umount", mountTarget)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Cvmfs mount failed.\nOutput:%s\nError%s\n",
			out, err)
	}

	return nil
}

func (cm *cvmfsManager) isConfigured(repo string) error {
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

func (cm *cvmfsManager) Get(repo string) error {
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

func (cm *cvmfsManager) Put(repo string) error {
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
	for k := range cm.ctr {
		cm.umount(k)
		cm.ctr[k] = 0
	}
	return nil
}

func (cm *cvmfsManager) GetLayers(layers ...ThinImageLayer) error {
	for i, l := range layers {
		if err := cm.Get(l.Repo); err != nil {
			cm.PutLayers(layers[:i]...)
			return err
		}
	}
	return nil
}

func (cm *cvmfsManager) Remount(repo string) error {
	if _, ok := cm.ctr[repo]; !ok {
		return nil
	}

	cmd := "cvmfs_talk"
	args := []string{"-i", repo, "remount", "sync"}

	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
		return err
	}

	return nil
}

func (cm *cvmfsManager) PutLayers(layers ...ThinImageLayer) error {
	for _, l := range layers {
		cm.Put(l.Repo)
	}
	return nil
}
