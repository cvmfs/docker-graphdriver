package main

import (
	// err "errors"
	"fmt"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/go-plugins-helpers/graphdriver"
	"io"
)

type MyDriver struct{}

func (d MyDriver) Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) error {
	fmt.Println("Init")
	return nil
}

func (d MyDriver) Create(id, parent, mountlabel string, storageOpt map[string]string) error {
	return nil
}

func (d MyDriver) CreateReadWrite(id, parent, mountlabel string, storageOpt map[string]string) error {
	return nil
}

func (d MyDriver) Remove(id string) error {
	return nil
}

func (d MyDriver) Get(id, mountLabel string) (string, error) {
	return "", nil
}

func (d MyDriver) Put(id string) error {
	return nil
}

func (d MyDriver) Exists(id string) bool {
	return true
}

func (d MyDriver) Status() [][2]string {
	var ret [][2]string
	return ret
}

func (d MyDriver) GetMetadata(id string) (map[string]string, error) {
	var ret map[string]string
	return ret, nil
}

func (d MyDriver) Cleanup() error {
	return nil
}

func (d MyDriver) Diff(id, parent string) io.ReadCloser {
	return nil
}

func (d MyDriver) Changes(id, parent string) ([]graphdriver.Change, error) {
	return nil, nil
}

func (d MyDriver) ApplyDiff(id, parent string, archive io.Reader) (int64, error) {
	return 0, nil
}

func (d MyDriver) DiffSize(id, parent string) (int64, error) {
	return 0, nil
}

func main() {
	d := MyDriver{}
	h := graphdriver.NewHandler(d)
	h.ServeUnix("test_graph", 0)
}
