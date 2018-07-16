package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	user, pass string
)

func init() {
	downloadManifestCmd.Flags().StringVarP(&user, "username", "u", "", "username to use to log in into the registry.")
	downloadManifestCmd.Flags().StringVarP(&pass, "password", "p", "", "password to use to log in into the registry.")
	rootCmd.AddCommand(downloadManifestCmd)
}

var downloadManifestCmd = &cobra.Command{
	Use:   "download-manifest",
	Short: "Download the manifest of the image, if sucessful it will print the manifest itself, otherwise will show what went wrong.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		img, err := lib.ParseImage(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if img.Tag == "" && img.Digest == "" {
			log.Fatal("Please provide either the image tag or the image digest")
		}
		url := img.GetManifestUrl()
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 401 {
			log.WithFields(log.Fields{
				"status code": resp.StatusCode,
			}).Info("Expected status code 401, print body anyway.")
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				lib.LogE(err).Fatal("Error in reading the first http response")
			}
			fmt.Println(string(body))
			os.Exit(1)
		}
		WwwAuthenticate := resp.Header["Www-Authenticate"][0]
		token, err := requestAuthToken(WwwAuthenticate)
		if err != nil {
			lib.LogE(err).Fatal("Error in getting the authentication token")
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to create a http request")
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

		resp, err = client.Do(req)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lib.LogE(err).Fatal("Error in reading the second http response")
		}
		fmt.Println(string(body))
	},
}

func parseBearerToken(token string) (realm string, options map[string]string, err error) {
	options = make(map[string]string)
	args := token[7:]
	keyValue := strings.Split(args, ",")
	for _, kv := range keyValue {
		splitted := strings.Split(kv, "=")
		if len(splitted) != 2 {
			err = fmt.Errorf("Wrong formatting of the token")
			return
		}
		splitted[1] = strings.Trim(splitted[1], `"`)
		if splitted[0] == "realm" {
			realm = splitted[1]
		} else {
			options[splitted[0]] = splitted[1]
		}
	}
	return
}

func requestAuthToken(token string) (authToken string, err error) {
	realm, options, err := parseBearerToken(token)
	if err != nil {
		return
	}
	req, err := http.NewRequest("GET", realm, nil)
	if err != nil {
		return
	}

	query := req.URL.Query()
	for k, v := range options {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = fmt.Errorf("Authorization error %s", resp.Status)
		return
	}

	var jsonResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}
	authTokenInterface, ok := jsonResp["token"]
	if ok {
		authToken = "Bearer " + authTokenInterface.(string)
		return
	} else {
		err = fmt.Errorf("Didn't get the token key from the server")
	}
	return
}
