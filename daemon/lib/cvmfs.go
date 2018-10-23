package lib

import (
	"os"
	"path/filepath"

	copy "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
)

// ingest into the repository, inside the subpath, the target (directory or file) object
// CVMFSRepo: just the name of the repository (ex: unpacked.cern.ch)
// path: the path inside the repository, without the prefix (ex: .foo/bar/baz)
// target: the path of the target in the FS
// if no error is returned, we remove the target from the FS
func IngestIntoCVMFS(CVMFSRepo string, path string, target string) (err error) {
	defer func() {
		if err == nil {
			Log().WithFields(log.Fields{"target": target, "action": "ingesting"}).Info("Deleting temporary directory")
			os.RemoveAll(target)
		}
	}()
	Log().WithFields(log.Fields{"target": target, "action": "ingesting"}).Info("Start ingesting")

	path = filepath.Join("/", "cvmfs", CVMFSRepo, path)

	Log().WithFields(log.Fields{"target": target, "action": "ingesting"}).Info("Start transaction")
	err = ExecCommand("cvmfs_server", "transaction", CVMFSRepo)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Error("Error in opening the transaction")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}

	Log().WithFields(log.Fields{"target": target, "action": "ingesting"}).Info("Copying directory")

	os.RemoveAll(path)
	err = os.MkdirAll(path, 0666)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Warning("Error in creating the directory where to copy the singularity")
	}

	err = copy.Copy(target, path)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo, "target": target}).Error("Error in moving the target inside the CVMFS repo")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}

	Log().WithFields(log.Fields{"target": target, "action": "ingesting"}).Info("Publishing")
	err = ExecCommand("cvmfs_server", "publish", CVMFSRepo)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Error("Error in publishing the repository")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}
	err = nil
	return err
}
