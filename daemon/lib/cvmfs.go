package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	copy "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
)

// ingest into the repository, inside the subpath, the target (directory or file) object
// CVMFSRepo: just the name of the repository (ex: unpacked.cern.ch)
// path: the path inside the repository, without the prefix (ex: .foo/bar/baz), where to put the ingested target
// target: the path of the target in the normal FS, the thing to ingest
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

	Log().WithFields(log.Fields{"target": target, "path": path, "action": "ingesting"}).Info("Copying target into path")

	targetStat, err := os.Stat(target)
	if err != nil {
		LogE(err).WithFields(log.Fields{"target": target}).Error("Impossible to obtain information about the target")
		return err
	}

	if targetStat.Mode().IsDir() {
		os.RemoveAll(path)
		err = os.MkdirAll(path, 0666)
		if err != nil {
			LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Warning("Error in creating the directory where to copy the singularity")
		}
		err = copy.Copy(target, path)

	} else if targetStat.Mode().IsRegular() {
		err = func() error {
			os.MkdirAll(filepath.Dir(path), 0666)
			os.Remove(path)
			from, err := os.Open(target)
			defer from.Close()
			if err != nil {
				return err
			}
			to, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
			defer to.Close()
			if err != nil {
				return err
			}
			_, err = io.Copy(to, from)
			return err
		}()
	} else {
		err = fmt.Errorf("Trying to ingest neither a file nor a directory")
	}

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

func CreateSymlinkIntoCVMFS(CVMFSRepo, newLinkName, toLinkPath string) (err error) {
	// check that we are creating a link inside the repository towards something in the repository
	dirsNew := strings.Split(newLinkName, string(filepath.Separator))
	if len(dirsNew) <= 3 || dirsNew[1] != "cvmfs" || dirsNew[2] != CVMFSRepo {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo, "linkName": newLinkName}).Error(
			"Error in creating the symlink, new outside repository")
		return fmt.Errorf("Trying to create a symlink outside the repository")
	}
	dirsToLink := strings.Split(toLinkPath, string(filepath.Separator))
	if len(dirsToLink) <= 3 || dirsNew[1] != "cvmfs" || dirsNew[2] != CVMFSRepo {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo, "toFile": toLinkPath}).Error(
			"Error in creating the symlink, trying to link to something outside the repository")
		return fmt.Errorf("Trying to link to something outside the repository")
	}
	// check if the file we want to link actually exists
	if _, err := os.Stat(toLinkPath); os.IsNotExist(err) {
		return err
	}

	err = ExecCommand("cvmfs_server", "transaction", CVMFSRepo)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Error("Error in opening the transaction")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}

	err = os.Symlink(newLinkName, toLinkPath)
	if err != nil {
		LogE(err).WithFields(log.Fields{
			"repo":     CVMFSRepo,
			"linkName": newLinkName,
			"toFile":   toLinkPath}).Error(
			"Error in creating the symlink")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}

	err = ExecCommand("cvmfs_server", "publish", CVMFSRepo)
	if err != nil {
		LogE(err).WithFields(log.Fields{"repo": CVMFSRepo}).Error("Error in publishing the repository")
		ExecCommand("cvmfs_server", "abort", "-f", CVMFSRepo)
		return err
	}
	return nil
}

func SaveLayersBacklink(CVMFSRepo string, img Image, layerMetadataPaths []string) error {
	llog := func(l *log.Entry) *log.Entry {
		return l.WithFields(log.Fields{"action": "save backlink",
			"repo":  CVMFSRepo,
			"image": img.GetSimpleName()})
	}
	type Backlink struct {
		Origin []string `json:"origin"`
	}
	llog(Log()).Info("Start saving backlinks")
	llog(Log()).Info("Start transaction")

	backlinks := make(map[string][]byte)

	for _, layerMetadataPath := range layerMetadataPaths {
		originPath := filepath.Join("/", "cvmfs", CVMFSRepo, layerMetadataPath, "origin.json")
		imgManifest, err := img.GetManifest()
		if err != nil {
			llog(LogE(err)).WithFields(log.Fields{"file": originPath}).Error(
				"Error in getting the manifest from the image, skipping...")
			continue
		}
		imgDigest := imgManifest.Config.Digest

		var backlink Backlink
		if _, err := os.Stat(originPath); os.IsNotExist(err) {

			backlink = Backlink{Origin: []string{imgDigest}}

		} else {

			backlinkFile, err := os.Open(originPath)
			if err != nil {
				llog(LogE(err)).WithFields(log.Fields{"file": originPath}).Error(
					"Error in opening the file for writing the backlinks, skipping...")
				continue
			}

			byteBackLink, err := ioutil.ReadAll(backlinkFile)
			if err != nil {
				llog(LogE(err)).WithFields(log.Fields{"file": originPath}).Error(
					"Error in reading the bytes from the origin file, skipping...")
				continue
			}

			err = json.Unmarshal(byteBackLink, &backlink)
			if err != nil {
				llog(LogE(err)).WithFields(log.Fields{"file": originPath}).Error(
					"Error in unmarshaling the files, skipping...")
				continue
			}

			backlink.Origin = append(backlink.Origin, imgDigest)
		}
		backlinkBytesMarshal, err := json.Marshal(backlink)
		if err != nil {
			llog(LogE(err)).WithFields(log.Fields{"file": originPath}).Error(
				"Error in Marshaling back the files, skipping...")
			continue
		}

		backlinks[originPath] = backlinkBytesMarshal
	}

	err := ExecCommand("cvmfs_server", "transaction", CVMFSRepo)
	if err != nil {
		llog(LogE(err)).Error("Error in opening the transaction")
		return err
	}

	for path, fileContent := range backlinks {
		// the path may not be there, check, and if it doesn't exists create it
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0666)
			if err != nil {
				llog(LogE(err)).WithFields(log.Fields{"file": path}).Error(
					"Error in creating the directory for the backlinks file, skipping...")
				continue
			}
		}
		err = ioutil.WriteFile(path, fileContent, 0666)
		if err != nil {
			llog(LogE(err)).WithFields(log.Fields{"file": path}).Error(
				"Error in writing the backlink file, skipping...")
			continue
		}
		llog(LogE(err)).WithFields(log.Fields{"file": path}).Info(
			"Wrote backlink")
	}

	err = ExecCommand("cvmfs_server", "publish", CVMFSRepo)
	if err != nil {
		llog(LogE(err)).Error("Error in publishing after adding the backlinks")
		return err
	}

	return nil
}
