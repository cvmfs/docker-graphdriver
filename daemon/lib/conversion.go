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
	"os/signal"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-errors/errors"
	log "github.com/sirupsen/logrus"

	d2c "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
)

var subDirInsideRepo = ".layers"

func ConvertWish(wish WishFriendly, convertAgain, forceDownload bool) (err error) {
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
			Log().WithFields(log.Fields{"layer": layer.Name}).Info("Start Ingesting routine")
			uncompressed, err := gzip.NewReader(layer.Resp.Body)
			defer layer.Resp.Body.Close()
			if err != nil {
				LogE(err).Error("Error in uncompressing the layer")
				noErrors = false
			}

			tempFile, err := ioutil.TempFile("", layer.Name)
			if err != nil {
				LogE(err).Error("Error in creating temporary file for storing the layer.")
				noErrors = false
				return
			}
			if err = tempFile.Chmod(0666); err != nil {
				LogE(err).Warning("Error in changing the mod of the temp file")
			}
			defer os.Remove(tempFile.Name())

			Log().WithFields(log.Fields{
				"layer": layer.Name,
				"file":  tempFile.Name(),
			}).Info("Copying the layer to a temporary file")
			if _, err = io.Copy(tempFile, uncompressed); err != nil {
				LogE(err).Error("Error in writing the layer into the temp file")
				noErrors = false
				return
			}

			// let's make sure that the file is actually written on disk
			tempFile.Sync()

			Log().Info("Finished copying the layer to the temporary file")

			if err = tempFile.Close(); err != nil {
				LogE(err).Warning("Error in closing the temp file where we wrote the layer")
			}

			Log().WithFields(log.Fields{"layer": layer.Name}).Info("Start Ingesting the file into CVMFS")
			layerDigest := strings.Split(layer.Name, ":")[1]
			layerLocation := subDirInsideRepo + "/" + layerDigest

			err = ExecCommand("cvmfs_server", "ingest", "-t", tempFile.Name(), "-b", layerLocation, wish.CvmfsRepo)
			if err != nil {
				LogE(err).WithFields(log.Fields{"layer": layer.Name}).Error("Some error in ingest the layer")
				noErrors = false
				cleanup(layerLocation)
				return
			}
		}
		Log().Info("Finished pushing the layers into CVMFS")
	}()
	// this wil start to feed the above goroutine by writing into layersChanell
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

	if noErrorInConversionValue {
		err = AddConverted(wish.Id, manifest)
		if err != nil && convertAgain == false {
			LogE(err).Error("Error in storing the conversion in the database")
		} else {
			Log().Info("Conversion completed")
		}
	} else {
		Log().Warn("Some error during the conversion, we are not storing it into the database")
	}
	return nil
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	if buf == nil {
		buf = make([]byte, size)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				LogE(ew).Error("Write error")
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
				LogE(err).Error("Read error")
				fmt.Println(err.(*errors.Error).ErrorStack())
			}
			break
		}
	}
	return written, err
}
