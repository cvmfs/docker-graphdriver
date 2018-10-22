package lib

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"

	d2c "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
)

var subDirInsideRepo = ".layers"

func ConvertWish(wish WishFriendly, convertAgain, forceDownload, convertSingularity bool) (err error) {
	interruptLayerUpload := make(chan os.Signal, 1)
	// register to Ctrl-C
	signal.Notify(interruptLayerUpload, os.Interrupt)

	outputImage, err := GetImageById(wish.OutputId)
	if err != nil {
		return
	}
	password, err := GetUserPassword(outputImage.User, outputImage.Registry)
	if err != nil {
		return
	}
	inputImage, err := GetImageById(wish.InputId)
	if err != nil {
		return
	}
	manifest, err := inputImage.GetManifest()
	if err != nil {
		return
	}
	if AlreadyConverted(wish.Id, manifest.Config.Digest) && convertAgain == false {
		Log().Info("Already converted the image, skipping.")
		return nil
	}
	layersChanell := make(chan downloadedLayer, 3)
	stopGettingLayers := make(chan bool, 1)
	noErrorInConversion := make(chan bool, 1)
	go func() {
		noErrors := true
		defer func() {
			noErrorInConversion <- noErrors
			stopGettingLayers <- true
			close(stopGettingLayers)
		}()
		cleanup := func(location string) {
			Log().Info("Running clean up function deleting the last layer.")

			err := ExecCommand("cvmfs_server", "abort", "-f", wish.CvmfsRepo)
			if err != nil {
				LogE(err).Warning("Error in the abort command inside the cleanup function, this warning is usually normal")
			}

			err = ExecCommand("cvmfs_server", "ingest", "--delete", location, wish.CvmfsRepo)
			if err != nil {
				LogE(err).Error("Error in the cleanup command")
			}
		}
		for layer := range layersChanell {
			select {
			case _, _ = <-interruptLayerUpload:
				{
					Log().Info("Received SIGINT, exiting")
					return
				}
			default:
				{
				}
			}
			defer os.Remove(layer.Path)

			Log().WithFields(log.Fields{"layer": layer.Name}).Info("Start Ingesting the file into CVMFS")
			layerLocation := subDirInsideRepo + "/" + layer.Name

			err = ExecCommand("cvmfs_server", "ingest", "-t", layer.Path, "-b", layerLocation, wish.CvmfsRepo)
			if err != nil {
				LogE(err).WithFields(log.Fields{"layer": layer.Name}).Error("Some error in ingest the layer")
				noErrors = false
				cleanup(layerLocation)
				return
			}
			Log().WithFields(log.Fields{"layer": layer.Name}).Info("Finish Ingesting the file")
		}
		Log().Info("Finished pushing the layers into CVMFS")
	}()
	// this wil start to feed the above goroutine by writing into layersChanell
	err = inputImage.GetLayers(layersChanell)

	var singularity Singularity
	if convertSingularity {
		singularity, err = inputImage.DownloadSingularityDirectory()
		if err != nil {
			LogE(err).Error("Error in dowloading the singularity image")
			return
		}
	}
	changes, _ := inputImage.GetChanges()

	repoLocation := fmt.Sprintf("%s/%s", wish.CvmfsRepo, subDirInsideRepo)
	thin := d2c.MakeThinImage(manifest, repoLocation, inputImage.WholeName())
	if err != nil {
		return
	}

	thinJson, err := json.MarshalIndent(thin, "", "  ")
	if err != nil {
		return
	}
	fmt.Println(string(thinJson))
	var imageTar bytes.Buffer
	tarFile := tar.NewWriter(&imageTar)
	header := &tar.Header{Name: "thin.json", Mode: 0644, Size: int64(len(thinJson))}
	err = tarFile.WriteHeader(header)
	if err != nil {
		return
	}
	_, err = tarFile.Write(thinJson)
	if err != nil {
		return
	}
	err = tarFile.Close()
	if err != nil {
		return
	}

	dockerClient, err := client.NewClientWithOpts(client.WithVersion("1.19"))
	if err != nil {
		return
	}

	image := types.ImageImportSource{
		Source:     bytes.NewBuffer(imageTar.Bytes()),
		SourceName: "-",
	}
	importOptions := types.ImageImportOptions{
		Tag:     outputImage.Tag,
		Message: "",
		Changes: changes,
	}
	importResult, err := dockerClient.ImageImport(
		context.Background(),
		image,
		outputImage.GetSimpleName(),
		importOptions)
	if err != nil {
		LogE(err).Error("Error in image import")
		return
	}
	defer importResult.Close()
	Log().Info("Created the image in the local docker daemon")

	// is necessary this mechanism to pass the authentication to the
	// dockers even if the documentation says otherwise
	authStruct := struct {
		Username string
		Password string
	}{
		Username: outputImage.User,
		Password: password,
	}
	authBytes, _ := json.Marshal(authStruct)
	authCredential := base64.StdEncoding.EncodeToString(authBytes)
	pushOptions := types.ImagePushOptions{
		RegistryAuth: authCredential,
	}

	res, err := dockerClient.ImagePush(
		context.Background(),
		outputImage.GetSimpleName(),
		pushOptions)
	if err != nil {
		return
	}
	// here is possible to use the result of the above ReadAll to have
	// informantion about the status of the upload.
	_, err = ioutil.ReadAll(res)
	if err != nil {
		return
	}
	Log().Info("Finish pushing the image to the registry")
	// we wait for the goroutines to finish
	// and if there was no error we add everything to the converted table
	noErrorInConversionValue := <-noErrorInConversion

	// here we can launch the ingestion for the singularity image
	if convertSingularity {
		err = singularity.IngestIntoCVMFS(wish.CvmfsRepo)
		if err != nil {
			LogE(err).Error("Error in ingesting the singularity image into the CVMFS repository")
			return
		}
	}

	if noErrorInConversionValue {
		err = AddConverted(wish.Id, manifest)
		if err != nil && convertAgain == false {
			LogE(err).Error("Error in storing the conversion in the database")
		} else {
			Log().Info("Conversion completed")
		}
		return
	} else {
		Log().Warn("Some error during the conversion, we are not storing it into the database")
		return
	}
}
