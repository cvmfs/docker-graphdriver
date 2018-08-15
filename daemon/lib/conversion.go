package lib

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"

	d2c "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
)

func ConvertWish(wish WishFriendly, convertAgain, forceDownload bool) (err error) {
	interruptLayerUpload := make(chan os.Signal, 1)
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
	fmt.Println(manifest)
	layersChanell := make(chan downloadedLayer, 3)
	stopGettingLayers := make(chan bool, 1)
	noErrorInConversion := make(chan bool, 1)
	subDirInsideRepo := "layers"
	go func() {
		noErrors := true
		defer func() {
			noErrorInConversion <- noErrors
			stopGettingLayers <- true
			close(stopGettingLayers)
		}()
		cleanup := func(location string) {
			Log().Info("Running clean up function deleting the last layer.")

			cmd := exec.Command("cvmfs_server", "abort", wish.CvmfsRepo)

			stdout, _ := cmd.StdoutPipe()

			stderr, _ := cmd.StderrPipe()

			Log().Info("Running abort")
			err = cmd.Start()
			if err != nil {
				LogE(err).Error("Error in starting the abort command inside the cleanup function")
			}

			slurpOut, _ := ioutil.ReadAll(stdout)
			Log().WithFields(log.Fields{"pipe": "STDOUT Abort"}).Info(string(slurpOut))

			slurpErr, err := ioutil.ReadAll(stderr)
			Log().WithFields(log.Fields{"pipe": "STDERR Abort"}).Info(string(slurpErr))

			err = cmd.Wait()
			if err != nil {
				LogE(err).Error("Error in the abort command inside the cleanup function")
			}

			cmd = exec.Command("cvmfs_server", "ingest", "--delete", location, wish.CvmfsRepo)
			stdout, _ = cmd.StdoutPipe()

			stderr, _ = cmd.StderrPipe()

			Log().Info("Running delete")
			err = cmd.Start()

			if err != nil {
				LogE(err).Error("Impossible to start the clean up command")
			}

			slurpOut, _ = ioutil.ReadAll(stdout)
			Log().WithFields(log.Fields{"pipe": "STDOUT Cleanup"}).Info(string(slurpOut))

			slurpErr, err = ioutil.ReadAll(stderr)
			Log().WithFields(log.Fields{"pipe": "STDERR Cleanup"}).Info(string(slurpErr))

			err = cmd.Wait()
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
			Log().WithFields(log.Fields{"layer": layer.Name}).Info("Ingesting layer")
			uncompressed, err := gzip.NewReader(layer.Resp.Body)
			defer layer.Resp.Body.Close()
			if err != nil {
				LogE(err).Error("Error in uncompressing the layer")
				noErrors = false
			}
			layerDigest := strings.Split(layer.Name, ":")[1]
			layerLocation := subDirInsideRepo + "/" + layerDigest
			cmd := exec.Command("cvmfs_server", "ingest", "-t", "-", "-b", layerLocation, wish.CvmfsRepo)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				noErrors = false
				return
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				noErrors = false
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				noErrors = false
				return
			}
			err = cmd.Start()

			go func() {
				defer stdin.Close()
				_, err := io.Copy(stdin, uncompressed)
				if err != nil {
					LogE(err).Error("Error in writing to stdin")
					noErrors = false
				}
			}()
			if err != nil {
				LogE(err).Error("Error in starting the ingest command")
				noErrors = false
				cleanup(layerLocation)
				return
			}

			slurpOut, err := ioutil.ReadAll(stdout)

			slurpErr, err := ioutil.ReadAll(stderr)

			err = cmd.Wait()
			if err != nil {
				LogE(err).WithFields(log.Fields{"layer": layer.Name}).Error("Some error in ingest the layer")
				fmt.Println("STDOUT: ", string(slurpOut))
				fmt.Println("STDERR: ", string(slurpErr))
				noErrors = false
				cleanup(layerLocation)
				return
			}
		}
		Log().Info("Finished pushing the layers into CVMFS")
	}()
	if forceDownload {
		err = inputImage.GetLayers(wish.CvmfsRepo, subDirInsideRepo, layersChanell, stopGettingLayers)
	} else {
		err = inputImage.GetLayerIfNotInCVMFS(wish.CvmfsRepo, subDirInsideRepo, layersChanell, stopGettingLayers)
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

	dockerClient, err := client.NewClientWithOpts()
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
	result, err := ioutil.ReadAll(res)
	if err != nil {
		return
	}
	fmt.Println(string(result))
	Log().Info("Finish pushing the image to the registry")
	// we wait for the goroutines to finish
	// and if there was no error we add everything to the converted table
	if <-noErrorInConversion {
		inputDigest := manifest.Config.Digest
		err = AddConverted(wish.Id, inputDigest)
		if err != nil && convertAgain == false {
			LogE(err).Error("Error in storing the conversion in the database")
		}
	} else {
		Log().Warn("Some error during the conversion, we are not storing it into the database")
	}
	return nil
}
