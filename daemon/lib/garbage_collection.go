package lib

import (
	"fmt"
	"path/filepath"
	"strings"
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
			err = ExecCommand("cvmfs_server", "ingest", "--delete", layer, repoName)
			if err != nil {
				LogE(err).Warning("Error in deleting the layer")
			}
		}
	}
	return nil
}
