package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	da "github.com/cvmfs/docker-graphdriver/daemon/docker-api"
)

func getAllLayers() []string {
	path := filepath.Join("/", "cvmfs", "*", subDirInsideRepo, "*")
	paths, _ := filepath.Glob(path)
	return paths
}

// this function is based on an "homegrow" implementation of a set intersection, go, since it does not provide generics, does not provides also generic datastructures like sets
func findToRemoveLayers(all_layers, needed_layers []string) (removes []string) {
	all := map[string]bool{}
	needed := map[string]bool{}
	to_remove := map[string]bool{}

	for _, layer := range all_layers {
		all[layer] = true
	}
	for _, layer := range needed_layers {
		needed[layer] = true
	}
	for layer, _ := range all {
		_, isNeeded := needed[layer]
		if !isNeeded {
			to_remove[layer] = true
		}
	}
	for layer := range to_remove {
		removes = append(removes, layer)
	}
	return
}

func findImageToGarbageCollect(CVMFSRepo string) ([]da.Manifest, error) {
	removeSchedulePath := RemoveScheduleLocation(CVMFSRepo)
	llog := func(l *log.Entry) *log.Entry {
		return l.WithFields(log.Fields{
			"action": "find image to garbage collect in schedule file",
			"file":   removeSchedulePath})
	}

	var schedule []da.Manifest

	_, err := os.Stat(removeSchedulePath)
	if os.IsNotExist(err) {
		return schedule, nil
	}
	if err != nil {
		llog(LogE(err)).Error("Error in stating the schedule file")
		return schedule, err
	}
	scheduleFileRO, err := os.Open(removeSchedulePath)
	if err != nil {
		llog(LogE(err)).Error("Error in opening the schedule file")
		return schedule, err
	}

	scheduleBytes, err := ioutil.ReadAll(scheduleFileRO)
	if err != nil {
		llog(LogE(err)).Error("Impossible to read the schedule file")
		return schedule, err
	}

	err = scheduleFileRO.Close()
	if err != nil {
		llog(LogE(err)).Error("Impossible to close the schedule file")
		return schedule, err
	}

	err = json.Unmarshal(scheduleBytes, &schedule)
	if err != nil {
		llog(LogE(err)).Error("Impossible to unmarshal the schedule file")
		return schedule, err
	}

	return schedule, nil
}

// with image and layer we pass the digest of the layer and the digest of the image, both without the sha256: prefix
func garbabeCollectSingleLayer(CVMFSRepo, image, layer string) error {
	backlink, err := getBacklinkFromLayer(CVMFSRepo, layer)
	llog := func(l *log.Entry) *log.Entry {
		return l.WithFields(log.Fields{"action": "garbage collect layer",
			"repo":  CVMFSRepo,
			"image": image,
			"layer": layer})
	}
	if err != nil {
		llog(LogE(err)).Error("Impossible to retrieve the backlink information")
		return err
	}
	var newOrigin []string
	for _, origin := range backlink.Origin {
		withoutPrefix := strings.Split(origin, ":")[1]
		if withoutPrefix != image {
			newOrigin = append(newOrigin, origin)
		}
	}
	if len(newOrigin) > 0 {
		backlink.Origin = newOrigin
		backLinkMarshall, err := json.Marshal(backlink)
		if err != nil {
			llog(LogE(err)).Error("Error in marshaling the new backlink")
			return err
		}

		backlinkPath := getBacklinkPath(CVMFSRepo, layer)

		err = ExecCommand("cvmfs_server", "transaction", CVMFSRepo).Start()
		if err != nil {
			llog(LogE(err)).Error("Error in opening the transaction")
			return err
		}

		dir := filepath.Dir(backlinkPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0666)
			if err != nil {
				llog(LogE(err)).WithFields(log.Fields{"directory": dir}).Error(
					"Error in creating the directory for the backlinks file, skipping...")
				return err
			}
		}

		err = ioutil.WriteFile(backlinkPath, backLinkMarshall, 0666)
		if err != nil {
			llog(LogE(err)).WithFields(log.Fields{"file": backlinkPath}).Error(
				"Error in writing the backlink file, skipping...")
			return err
		}

		err = ExecCommand("cvmfs_server", "publish", CVMFSRepo).Start()
		if err != nil {
			llog(LogE(err)).Error("Error in publishing after adding the backlinks")
			return err
		}
		// write it to file
		return nil
	} else {
		err = RemoveLayer(CVMFSRepo, layer)
		if err != nil {
			llog(LogE(err)).Error("Error in deleting the layer")
		}
		return err
	}
}

func RemoveUselessLayers() error {
	all_layers := getAllLayers()
	needed_layers, err := GetAllNeededLayers()
	if err != nil {
		return err
	}
	to_remove := findToRemoveLayers(all_layers, needed_layers)
	// to_remove is now a slice of paths that we wish to remove from the
	// repository.
	// However those paths are "complete":
	// `/cvmfs/$repo.name/layers/fffbbbaaa`
	// we need to group them by the repository they belong to, and then
	// extract only the last part of the path we care about:
	// `layers/fffbbbaaa`
	paths := map[string][]string{}
	for _, path := range to_remove {
		pathSplitted := strings.Split(path, "/")
		repoName := pathSplitted[2]
		paths[repoName] = append(paths[repoName], strings.Join(pathSplitted[3:], "/"))
	}
	fmt.Println(paths)
	// possible optimization
	// batch the delete of layers, not all together that it will lock the repo for quite too long
	// but few of them like 10 or 20 together is something very reasonable.
	for repoName, layers := range paths {
		for _, layer := range layers {
			err = ExecCommand("cvmfs_server", "ingest", "--delete", layer, repoName).Start()
			if err != nil {
				LogE(err).Warning("Error in deleting the layer")
			}
		}
	}
	return nil
}
