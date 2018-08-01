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
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"

	d2c "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
)

func ConvertDesiderata(des DesiderataFriendly, convertAgain, forceDownload bool) (err error) {
	outputImage, err := GetImageById(des.OutputId)
	if err != nil {
		return
	}
	password, err := GetUserPassword(outputImage.User, outputImage.Registry)
	if err != nil {
		return
	}
	inputImage, err := GetImageById(des.InputId)
	if err != nil {
		return
	}
	manifest, err := inputImage.GetManifest()
	if err != nil {
		return
	}
	if AlreadyConverted(des.Id, manifest.Config.Digest) && convertAgain == false {
		Log().Info("Already converted the image, skipping.")
		return nil
	}
	fmt.Println(manifest)
	layersChanell := make(chan downloadedLayer, 3)
	noErrorInConversion := make(chan bool, 1)
	subDirInsideRepo := "layers"
	go func() {
		noErrors := true
		for layer := range layersChanell {
			uncompressed, err := gzip.NewReader(layer.Resp.Body)
			defer layer.Resp.Body.Close()
			if err != nil {
				LogE(err).Error("Error in uncompressing the layer")
			}
			layerDigest := strings.Split(layer.Name, ":")[1]
			cmd := exec.Command("cvmfs_server", "ingest", "-t", "-", "-b", subDirInsideRepo+"/"+layerDigest, des.CvmfsRepo)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				return
			}
			err = cmd.Start()

			go func() {
				_, err := io.Copy(stdin, uncompressed)
				if err != nil {
					fmt.Println("Error in writing to stdin: ", err)
				}
			}()
			if err != nil {
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
			}
		}
		Log().Info("Finished pushing the layers into CVMFS")
		noErrorInConversion <- noErrors
	}()
	if forceDownload {
		err = inputImage.GetLayers(des.CvmfsRepo, subDirInsideRepo, layersChanell)
	} else {
		err = inputImage.GetLayerIfNotInCVMFS(des.CvmfsRepo, subDirInsideRepo, layersChanell)
	}
	repoLocation := fmt.Sprintf("%s/%s", des.CvmfsRepo, subDirInsideRepo)
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
		Changes: []string{"ENV CVMFS_IMAGE true"},
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
		err = AddConverted(des.Id, inputDigest)
		if err != nil && convertAgain == false {
			LogE(err).Error("Error in storing the conversion in the database")
		}
	} else {
		Log().Warn("Some error during the conversion, we are not storing it into the database")
	}
	return nil
}
