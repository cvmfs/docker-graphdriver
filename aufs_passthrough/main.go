package main

import (
	// err "errors"
	"fmt"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/daemon/graphdriver/aufs"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/reexec"
	plugin "github.com/docker/go-plugins-helpers/graphdriver"
	"io"
)

var driver graphdriver.Driver

type MyDriver struct {
}

func (d MyDriver) Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) error {
	fmt.Printf("Init(%s)\n", home)
	var err error
	driver, err = aufs.Init(home, options, uidMaps, gidMaps)
	return err
}

func (d MyDriver) Create(id, parent, mountLabel string, storageOpt map[string]string) error {
	fmt.Printf("Create(%s, %s, %s)\n", id, parent, mountLabel)
	opts := graphdriver.CreateOpts{MountLabel: mountLabel, StorageOpt: storageOpt}
	return driver.Create(id, parent, &opts)
}

func (d MyDriver) CreateReadWrite(id, parent, mountLabel string, storageOpt map[string]string) error {
	fmt.Printf("CreateReadWrite(%s, %s, %s)\n", id, parent, mountLabel)
	opts := graphdriver.CreateOpts{MountLabel: mountLabel, StorageOpt: storageOpt}
	return driver.CreateReadWrite(id, parent, &opts)
}

func (d MyDriver) Remove(id string) error {
	fmt.Printf("Remove(%s)\n", id)
	return driver.Remove(id)
}

func (d MyDriver) Get(id, mountLabel string) (string, error) {
	fmt.Printf("Get(%s, %s)\n", id, mountLabel)
	return driver.Get(id, mountLabel)
}

func (d MyDriver) Put(id string) error {
	fmt.Printf("Put(%s)\n", id)
	return driver.Put(id)
}

func (d MyDriver) Exists(id string) bool {
	fmt.Printf("Exists(%s)\n", id)
	return driver.Exists(id)
}

func (d MyDriver) Status() [][2]string {
	fmt.Println("Status()")
	return driver.Status()
}

func (d MyDriver) GetMetadata(id string) (map[string]string, error) {
	fmt.Printf("GetMetadata(%s)\n", id)
	return driver.GetMetadata(id)
}

func (d MyDriver) Cleanup() error {
	fmt.Println("Cleanup")
	return driver.Cleanup()
}

func (d MyDriver) Diff(id, parent string) io.ReadCloser {
	fmt.Printf("Diff(%s, %s)\n", id, parent)
	closer, _ := driver.Diff(id, parent)
	return closer
}

func (d MyDriver) Changes(id, parent string) ([]plugin.Change, error) {
	fmt.Printf("Changes(%, %s)\n", id, parent)
	changes, err := driver.Changes(id, parent)
	ret := make([]plugin.Change, len(changes))

	for i, v := range changes {
		ret[i] = plugin.Change{Path: v.Path, Kind: plugin.ChangeKind(v.Kind)}
	}

	return ret, err
}

func (d MyDriver) ApplyDiff(id, parent string, archive io.Reader) (int64, error) {
	fmt.Printf("ApplyDiff(%s, %s)\n", id, parent)
	return driver.ApplyDiff(id, parent, archive)
}

func (d MyDriver) DiffSize(id, parent string) (int64, error) {
	fmt.Printf("DiffSize(%s, %s)\n", id, parent)
	return driver.DiffSize(id, parent)
}

func main() {
	if reexec.Init() {
		return
	}

	d := MyDriver{}
	h := plugin.NewHandler(d)
	h.ServeUnix("aufs_passthrough", 0)
}
