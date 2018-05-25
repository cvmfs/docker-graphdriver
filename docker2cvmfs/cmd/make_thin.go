package cmd

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var MakeThin = &cobra.Command{
	Use:   "make-thin creates a thin image out of a regular docker images storing the files inside the provided repository.",
	Short: "Directly creates a thin image from a regular docker image.",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		registry := flags.Lookup("registry").Value.String()
		inputReference := flags.Lookup("input-reference").Value.String()
		outputReference := flags.Lookup("output-reference").Value.String()
		repository := flags.Lookup("repository").Value.String()
		subdirectory := flags.Lookup("subdirectory").Value.String()
		err := lib.PullLayers(registry, inputReference, repository, subdirectory)
		if err != nil {
			log.Fatal(err)
		}
		manifest, err := lib.GetManifest(registry, inputReference)
		origin := inputReference + "@" + registry
		thinImage := lib.MakeThinImage(manifest, repository+"/"+strings.TrimSuffix(subdirectory, "/"), origin)
		thinImageJson, err := json.MarshalIndent(thinImage, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(thinImageJson))

		var imageTarFileStorange bytes.Buffer
		tarFile := tar.NewWriter(&imageTarFileStorange)
		header := &tar.Header{Name: "thin-image.json",
			Mode: 0600,
			Size: int64(len(thinImageJson)),
		}
		err = tarFile.WriteHeader(header)
		if err != nil {
			log.Fatal("Error in creating the tarfile for the thin image. [WriteHeader]", err)
		}
		_, err = tarFile.Write(thinImageJson)
		if err != nil {
			log.Fatal("Error in creating the tarfile for the thin image. [Write] ", err)
		}
		err = tarFile.Close()
		if err != nil {
			log.Fatal("Error in creating the tarfile for the thin image. [Close] ", err)
		}

		dockerClient, err := client.NewEnvClient()
		if err != nil {
			log.Fatal("Impossible to get a docker client using your env variables: ", err)
		}
		image := types.ImageImportSource{
			Source:     bytes.NewBuffer(imageTarFileStorange.Bytes()),
			SourceName: "-",
		}
		importResult, err := dockerClient.ImageImport(context.Background(), image, outputReference, types.ImageImportOptions{})
		if err != nil {
			log.Fatal("Error in importing the images: ", err)
		} else {
			defer importResult.Close()
		}
		importResultBuffer := new(bytes.Buffer)
		importResultBuffer.ReadFrom(importResult)
		fmt.Printf(importResultBuffer.String())
	},
}
