// +build linux

/*

aufs driver directory structure

  .
  ├── layers // Metadata of layers
  │   ├── 1
  │   ├── 2
  │   └── 3
  ├── diff  // Content of the layer
  │   ├── 1  // Contains layers that need to be mounted for the id
  │   ├── 2
  │   └── 3
  └── mnt    // Mount points for the rw layers to be mounted
      ├── 1
      ├── 2
      └── 3

*/

package aufs

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/vbatts/tar-split/tar/storage"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/directory"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/locker"
	mountpk "github.com/docker/docker/pkg/mount"

	"github.com/cvmfs/docker-graphdriver/plugins/util"
	"github.com/opencontainers/selinux/go-selinux/label"
)

var (
	// ErrAufsNotSupported is returned if aufs is not supported by the host.
	ErrAufsNotSupported = fmt.Errorf("AUFS was not found in /proc/filesystems")
	// ErrAufsNested means aufs cannot be used bc we are in a user namespace
	ErrAufsNested = fmt.Errorf("AUFS cannot be used in non-init user namespace")
	backingFs     = "<unknown>"

	enableDirpermLock sync.Once
	enableDirperm     bool
)

func init() {
	graphdriver.Register("aufs", Init)
}

// Driver contains information about the filesystem mounted.
type Driver struct {
	sync.Mutex
	root          string
	uidMaps       []idtools.IDMap
	gidMaps       []idtools.IDMap
	ctr           *graphdriver.RefCounter
	pathCacheLock sync.Mutex
	pathCache     map[string]string
	naiveDiff     graphdriver.DiffDriver
	locker        *locker.Locker
	cvmfsManager  util.ICvmfsManager

	cvmfsMountMethod string
	cvmfsMountPath   string
}

// Init returns a new AUFS driver.
// An error is returned if AUFS is not supported.
func Init(root string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	os.MkdirAll(root, os.ModePerm)
	os.MkdirAll("/dev/shm", os.ModePerm)
	exec.Command("mount", "-t", "tmpfs", "shmfs", "/dev/shm").Run()

	fsMagic, err := graphdriver.GetFSMagic(root)
	if err != nil {
		return nil, err
	}
	if fsName, ok := graphdriver.FsNames[fsMagic]; ok {
		backingFs = fsName
	}

	switch fsMagic {
	case graphdriver.FsMagicAufs, graphdriver.FsMagicBtrfs, graphdriver.FsMagicEcryptfs:
		logrus.Errorf("AUFS is not supported over %s", backingFs)
		return nil, graphdriver.ErrIncompatibleFS
	}

	paths := []string{
		"mnt",
		"diff",
		"layers",
	}

	a := &Driver{
		root:           root,
		uidMaps:        uidMaps,
		gidMaps:        gidMaps,
		pathCache:      make(map[string]string),
		ctr:            graphdriver.NewRefCounter(graphdriver.NewFsChecker(graphdriver.FsMagicAufs)),
		cvmfsMountPath: "/mnt/cvmfs",
	}

	if err := a.configureCvmfs(options); err != nil {
		return nil, err
	}

	a.cvmfsManager = util.NewCvmfsManager(a.cvmfsMountPath, a.cvmfsMountMethod)

	rootUID, rootGID, err := idtools.GetRootUIDGID(uidMaps, gidMaps)
	if err != nil {
		return nil, err
	}
	// Create the root aufs driver dir and return
	// if it already exists
	// If not populate the dir structure
	if err := idtools.MkdirAllAs(root, 0700, rootUID, rootGID); err != nil {
		if os.IsExist(err) {
			return a, nil
		}
		return nil, err
	}

	// if err := mountpk.MakePrivate(root); err != nil {
	//	return nil, err
	// }

	// Populate the dir structure
	for _, p := range paths {
		if err := idtools.MkdirAllAs(path.Join(root, p), 0700, rootUID, rootGID); err != nil {
			return nil, err
		}
	}

	a.naiveDiff = graphdriver.NewNaiveDiffDriver(a, uidMaps, gidMaps)
	return a, nil
}

func (a *Driver) rootPath() string {
	return a.root
}

func (*Driver) String() string {
	return "aufs"
}

// Status returns current information about the filesystem such as root directory, number of directories mounted, etc.
func (a *Driver) Status() [][2]string {
	ids, _ := loadIds(path.Join(a.rootPath(), "layers"))
	return [][2]string{
		{"Root Dir", a.rootPath()},
		{"Backing Filesystem", backingFs},
		{"Dirs", fmt.Sprintf("%d", len(ids))},
		{"Dirperm1 Supported", fmt.Sprintf("%v", useDirperm())},
	}
}

// GetMetadata not implemented
func (a *Driver) GetMetadata(id string) (map[string]string, error) {
	return nil, nil
}

// Exists returns true if the given id is registered with
// this driver
func (a *Driver) Exists(id string) bool {
	if _, err := os.Lstat(path.Join(a.rootPath(), "layers", id)); err != nil {
		return false
	}
	return true
}

// CreateReadWrite creates a layer that is writable for use as a container
// file system.
func (a *Driver) CreateReadWrite(id, parent string, opts *graphdriver.CreateOpts) error {
	fmt.Printf("CreateReadWrite(%s, %s)\n", id, parent)

	return a.Create(id, parent, opts)
}

// Create three folders for each id
// mnt, layers, and diff
func (a *Driver) Create(id, parent string, opts *graphdriver.CreateOpts) error {
	fmt.Printf("Create(%s, %s)\n", id, parent)

	if opts != nil && len(opts.StorageOpt) != 0 {
		return fmt.Errorf("--storage-opt is not supported for aufs")
	}

	if err := a.createDirsFor(id); err != nil {
		return err
	}
	// Write the layers metadata
	f, err := os.Create(path.Join(a.rootPath(), "layers", id))
	if err != nil {
		return err
	}
	defer f.Close()

	if parent != "" {
		ids, err := getParentIDs(a.rootPath(), parent)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(f, parent); err != nil {
			return err
		}
		for _, i := range ids {
			if _, err := fmt.Fprintln(f, i); err != nil {
				return err
			}
		}
	}

	return nil
}

// createDirsFor creates two directories for the given id.
// mnt and diff
func (a *Driver) createDirsFor(id string) error {
	paths := []string{
		"mnt",
		"diff",
	}

	rootUID, rootGID, err := idtools.GetRootUIDGID(a.uidMaps, a.gidMaps)
	if err != nil {
		return err
	}
	// Directory permission is 0755.
	// The path of directories are <aufs_root_path>/mnt/<image_id>
	// and <aufs_root_path>/diff/<image_id>
	for _, p := range paths {
		if err := idtools.MkdirAllAs(path.Join(a.rootPath(), p, id), 0755, rootUID, rootGID); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to debug EBUSY errors on remove.
func debugEBusy(mountPath string) (out []string, err error) {
	// lsof is not part of GNU coreutils. This is a best effort
	// attempt to detect offending processes.
	c := exec.Command("lsof")

	r, err := c.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Assigning pipes failed with %v", err)
	}

	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("Starting %s failed with %v", c.Path, err)
	}

	defer func() {
		waiterr := c.Wait()
		if waiterr != nil && err == nil {
			err = fmt.Errorf("Waiting for %s failed with %v", c.Path, waiterr)
		}
	}()

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		entry := sc.Text()
		if strings.Contains(entry, mountPath) {
			out = append(out, entry, "\n")
		}
	}

	return out, nil
}

// Remove will unmount and remove the given id.
func (a *Driver) Remove(id string) error {
	fmt.Printf("Remove(%s)\n", id)

	a.pathCacheLock.Lock()
	mountpoint, exists := a.pathCache[id]
	a.pathCacheLock.Unlock()
	if !exists {
		mountpoint = a.getMountpoint(id)
	}

	var retries int
	for {
		mounted, err := a.mounted(mountpoint)
		if err != nil {
			return err
		}
		if !mounted {
			break
		}

		if err := a.unmount(mountpoint); err != nil {
			if err != syscall.EBUSY {
				return fmt.Errorf("aufs: unmount error: %s: %v", mountpoint, err)
			}
			if retries >= 5 {
				out, debugErr := debugEBusy(mountpoint)
				if debugErr == nil {
					logrus.Warnf("debugEBusy returned %v", out)
				}
				return fmt.Errorf("aufs: unmount error after retries: %s: %v", mountpoint, err)
			}
			// If unmount returns EBUSY, it could be a transient error. Sleep and retry.
			retries++
			logrus.Warnf("unmount failed due to EBUSY: retry count: %d", retries)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	// Atomically remove each directory in turn by first moving it out of the
	// way (so that docker doesn't find it anymore) before doing removal of
	// the whole tree.
	tmpMntPath := path.Join(a.mntPath(), fmt.Sprintf("%s-removing", id))
	if err := os.Rename(mountpoint, tmpMntPath); err != nil && !os.IsNotExist(err) {
		if err == syscall.EBUSY {
			logrus.Warn("os.Rename err due to EBUSY")
			out, debugErr := debugEBusy(mountpoint)
			if debugErr == nil {
				logrus.Warnf("debugEBusy returned %v", out)
			}
		}
		return err
	}
	defer os.RemoveAll(tmpMntPath)

	tmpDiffpath := path.Join(a.diffPath(), fmt.Sprintf("%s-removing", id))
	if err := os.Rename(a.getDiffPath(id), tmpDiffpath); err != nil && !os.IsNotExist(err) {
		return err
	}
	defer os.RemoveAll(tmpDiffpath)

	// Remove the layers file for the id
	if err := os.Remove(path.Join(a.rootPath(), "layers", id)); err != nil && !os.IsNotExist(err) {
		return err
	}

	a.pathCacheLock.Lock()
	delete(a.pathCache, id)
	a.pathCacheLock.Unlock()
	return nil
}

// Get returns the rootfs path for the id.
// This will mount the dir at its given path
func (a *Driver) Get(id, mountLabel string) (string, error) {
	fmt.Printf("Get(%s, %s)\n", id, mountLabel)

	parents, err := a.getParentLayerPaths(id)
	if err != nil {
		return "", err
	}

	a.pathCacheLock.Lock()
	m, exists := a.pathCache[id]
	a.pathCacheLock.Unlock()

	if !exists {
		m = a.getDiffPath(id)
		if len(parents) > 0 {
			m = a.getMountpoint(id)
		}
	}
	if count := a.ctr.Increment(m); count > 1 {
		return m, nil
	}

	// If a dir does not have a parent ( no layers )do not try to mount
	// just return the diff path to the data
	if len(parents) > 0 {
		if err := a.mount(id, m, mountLabel, parents); err != nil {
			return "", err
		}
	}

	a.pathCacheLock.Lock()
	a.pathCache[id] = m
	a.pathCacheLock.Unlock()
	return m, nil
}

// Put unmounts and updates list of active mounts.
func (a *Driver) Put(id string) error {
	fmt.Printf("Put(%s)\n", id)

	a.pathCacheLock.Lock()
	m, exists := a.pathCache[id]
	if !exists {
		m = a.getMountpoint(id)
		a.pathCache[id] = m
	}
	a.pathCacheLock.Unlock()
	if count := a.ctr.Decrement(m); count > 0 {
		return nil
	}

	err := a.unmount(m)
	if err != nil {
		logrus.Debugf("Failed to unmount %s aufs: %v", id, err)
	}
	return err
}

// isParent returns if the passed in parent is the direct parent of the passed in layer
func (a *Driver) isParent(id, parent string) bool {
	fmt.Printf("isParent(%s, %s)\n", id, parent)

	parents, _ := getParentIDs(a.rootPath(), id)
	if parent == "" && len(parents) > 0 {
		return false
	}
	return !(len(parents) > 0 && parent != parents[0])
}

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be "".
func (a *Driver) Diff(id, parent string) (io.ReadCloser, error) {
	fmt.Printf("Diff(%s, %s)\n", id, parent)
	var newThinLayer string
	var exportPath string
	var isThin bool

	thin, err := a.getParentThinLayer(id)

	if err == nil {
		fmt.Println("Found a thin image!")
		isThin = true

		orig := a.getDiffPath(id)

		fmt.Printf("Orig diffpath: %s\n", orig)
		newLayer, err := a.cvmfsManager.UploadNewLayer(orig)
		if err != nil {
			fmt.Printf("error on UploadNewLayer(): %s", err.Error())
			return nil, err
		}

		fmt.Printf("Uploaded hash is: %s\n", newLayer.Digest)

		thin.AddLayer(newLayer)
		if newThinLayer, err = util.WriteThinFile(thin); err != nil {
			fmt.Printf("Failed to create thin file")
			return nil, err
		}
	} else {
		fmt.Println("Didn't find a thin image parent!")
		fmt.Println(err)
	}

	if isThin {
		exportPath = newThinLayer
	} else {
		exportPath = path.Join(a.rootPath(), "diff", id)
	}

	if !a.isParent(id, parent) {
		return a.naiveDiff.Diff(id, parent)
	}

	return archive.TarWithOptions(exportPath, &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: []string{archive.WhiteoutMetaPrefix + "*", "!" + archive.WhiteoutOpaqueDir},
		UIDMaps:         a.uidMaps,
		GIDMaps:         a.gidMaps,
	})
}

type fileGetNilCloser struct {
	storage.FileGetter
}

func (f fileGetNilCloser) Close() error {
	return nil
}

// DiffGetter returns a FileGetCloser that can read files from the directory that
// contains files for the layer differences. Used for direct access for tar-split.
func (a *Driver) DiffGetter(id string) (graphdriver.FileGetCloser, error) {
	fmt.Printf("DiffGetter(%s)\n", id)

	p := path.Join(a.rootPath(), "diff", id)
	return fileGetNilCloser{storage.NewPathFileGetter(p)}, nil
}

func (a *Driver) applyDiff(id string, diff io.Reader) error {
	fmt.Printf("applyDiff(%s)\n", id)

	return chrootarchive.UntarUncompressed(diff, path.Join(a.rootPath(), "diff", id), &archive.TarOptions{
		UIDMaps: a.uidMaps,
		GIDMaps: a.gidMaps,
	})
}

// DiffSize calculates the changes between the specified id
// and its parent and returns the size in bytes of the changes
// relative to its base filesystem directory.
func (a *Driver) DiffSize(id, parent string) (size int64, err error) {
	fmt.Printf("DiffSize(%s, %s)\n", id, parent)

	if !a.isParent(id, parent) {
		return a.naiveDiff.DiffSize(id, parent)
	}
	// AUFS doesn't need the parent layer to calculate the diff size.
	return directory.Size(path.Join(a.rootPath(), "diff", id))
}

// ApplyDiff extracts the changeset from the given diff into the
// layer with the specified id and parent, returning the size of the
// new layer in bytes.
func (a *Driver) ApplyDiff(id, parent string, diff io.Reader) (size int64, err error) {
	fmt.Printf("ApplyDiff(%s, %s)\n", id, parent)

	if !a.isParent(id, parent) {
		return a.naiveDiff.ApplyDiff(id, parent, diff)
	}

	// AUFS doesn't need the parent id to apply the diff if it is the direct parent.
	if err = a.applyDiff(id, diff); err != nil {
		return
	}

	return a.DiffSize(id, parent)
}

// Changes produces a list of changes between the specified layer
// and its parent layer. If parent is "", then all changes will be ADD changes.
func (a *Driver) Changes(id, parent string) ([]archive.Change, error) {
	fmt.Printf("Changes(%s, %s)\n", id, parent)

	if !a.isParent(id, parent) {
		return a.naiveDiff.Changes(id, parent)
	}

	// AUFS doesn't have snapshots, so we need to get changes from all parent
	// layers.
	layers, err := a.getParentLayerPaths(id)
	if err != nil {
		return nil, err
	}
	return archive.Changes(layers, path.Join(a.rootPath(), "diff", id))
}

func (a *Driver) getParentLayerPaths(id string) ([]string, error) {
	fmt.Printf("getParentLayerPaths(%s)\n", id)

	parentIds, err := getParentIDs(a.rootPath(), id)
	if err != nil {
		return nil, err
	}
	layers := make([]string, len(parentIds))
	foundThin := false
	ctr := 0

	// Get the diff paths for all the parent ids
	for _, p := range parentIds {
		diffPath := a.getDiffPath(p)

		if util.IsThinImageLayer(diffPath) && (foundThin == false) {
			nested_layers := util.GetNestedLayerIDs(diffPath)

			var err error
			if a.cvmfsMountMethod == "internal" {
				err = a.cvmfsManager.GetLayers(nested_layers...)
			}

			if err != nil {
				return nil, err
			}

			cvmfs_paths := util.GetCvmfsLayerPaths(nested_layers, a.cvmfsMountPath)
			layers = util.ExpandCvmfsLayerPaths(layers, cvmfs_paths, ctr)
			foundThin = true
			ctr += len(cvmfs_paths)
		} else if !util.IsThinImageLayer(diffPath) {
			layers[ctr] = path.Join(a.rootPath(), "diff", p)
			ctr += 1
		}
	}

	layers = layers[:ctr]
	return layers, nil
}

func (a *Driver) mount(id string, target string, mountLabel string, layers []string) error {
	a.Lock()
	defer a.Unlock()

	// If the id is mounted or we get an error return
	if mounted, err := a.mounted(target); err != nil || mounted {
		return err
	}

	rw := a.getDiffPath(id)

	if err := a.aufsMount(layers, rw, target, mountLabel); err != nil {
		return fmt.Errorf("error creating aufs mount to %s: %v", target, err)
	}
	return nil
}

func (a *Driver) unmount(mountPath string) error {
	a.Lock()
	defer a.Unlock()

	if mounted, err := a.mounted(mountPath); err != nil || !mounted {
		return err
	}
	if err := Unmount(mountPath); err != nil {
		return err
	}
	return nil
}

func (a *Driver) mounted(mountpoint string) (bool, error) {
	return graphdriver.Mounted(graphdriver.FsMagicAufs, mountpoint)
}

// Cleanup aufs and unmount all mountpoints
func (a *Driver) Cleanup() error {
	if a.cvmfsMountMethod == "internal" {
		a.cvmfsManager.PutAll()
	}

	exec.Command("umount", "/dev/shm").Run()

	var dirs []string
	if err := filepath.Walk(a.mntPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	}); err != nil {
		return err
	}

	for _, m := range dirs {
		if err := a.unmount(m); err != nil {
			logrus.Debugf("aufs error unmounting %s: %s", m, err)
		}
	}
	return mountpk.Unmount(a.root)
}

func (a *Driver) aufsMount(ro []string, rw, target, mountLabel string) (err error) {
	defer func() {
		if err != nil {
			Unmount(target)
		}
	}()

	// Mount options are clipped to page size(4096 bytes). If there are more
	// layers then these are remounted individually using append.

	offset := 54
	if useDirperm() {
		offset += len("dirperm1")
	}
	b := make([]byte, syscall.Getpagesize()-len(mountLabel)-offset) // room for xino & mountLabel
	bp := copy(b, fmt.Sprintf("br:%s=rw", rw))

	index := 0
	for ; index < len(ro); index++ {
		layer := fmt.Sprintf(":%s=ro+wh", ro[index])
		if bp+len(layer) > len(b) {
			break
		}
		bp += copy(b[bp:], layer)
	}

	opts := "dio,xino=/dev/shm/aufs.xino"
	if useDirperm() {
		opts += ",dirperm1"
	}
	data := label.FormatMountLabel(fmt.Sprintf("%s,%s", string(b[:bp]), opts), mountLabel)
	if err = mount("none", target, "aufs", 0, data); err != nil {
		return
	}

	for ; index < len(ro); index++ {
		layer := fmt.Sprintf(":%s=ro+wh", ro[index])
		data := label.FormatMountLabel(fmt.Sprintf("append%s", layer), mountLabel)
		if err = mount("none", target, "aufs", syscall.MS_REMOUNT, data); err != nil {
			return
		}
	}

	return
}

// useDirperm checks dirperm1 mount option can be used with the current
// version of aufs.
func useDirperm() bool {
	enableDirpermLock.Do(func() {
		base, err := ioutil.TempDir("", "docker-aufs-base")
		if err != nil {
			logrus.Errorf("error checking dirperm1: %v", err)
			return
		}
		defer os.RemoveAll(base)

		union, err := ioutil.TempDir("", "docker-aufs-union")
		if err != nil {
			logrus.Errorf("error checking dirperm1: %v", err)
			return
		}
		defer os.RemoveAll(union)

		opts := fmt.Sprintf("br:%s,dirperm1,xino=/dev/shm/aufs.xino", base)
		if err := mount("none", union, "aufs", 0, opts); err != nil {
			return
		}
		enableDirperm = true
		if err := Unmount(union); err != nil {
			logrus.Errorf("error checking dirperm1: failed to unmount %v", err)
		}
	})
	return enableDirperm
}

func (a *Driver) configureCvmfs(options []string) error {
	m, err := util.ParseOptions(options)

	if err != nil {
		return err
	}

	if method, ok := m["cvmfsMountMethod"]; !ok {
		a.cvmfsMountMethod = "internal"
	} else {
		a.cvmfsMountMethod = method
	}

	a.cvmfsMountPath = path.Join(a.root, "cvmfs")
	os.MkdirAll(a.cvmfsMountPath, os.ModePerm)

	return nil
}
