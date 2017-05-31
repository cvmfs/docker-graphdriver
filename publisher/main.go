package main

import "net/http"
import "fmt"
import "encoding/json"
import "bytes"
import "os/exec"
import "path"
import "io/ioutil"
import "os"

type StoredObjectRecord struct {
	S3 struct {
		Bucket struct {
			Name string
		}
		Object struct {
			Key string
		}
	}
}

type WebhookPayload struct {
	EventType string
	Records   []StoredObjectRecord
}

type Object struct {
	Bucket string
	Key    string
}

func (p *WebhookPayload) Object() Object {
	s3 := p.Records[0].S3

	return Object{
		Bucket: s3.Bucket.Name,
		Key:    s3.Object.Key,
	}
}

type PublisherConfig struct {
	CvmfsRepo        string
	MinioStoragePath string
}

var publisherConfig PublisherConfig

func LoadConfig(configPath string) (config PublisherConfig, err error) {
	var f []byte

	f, err = ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(f, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

type CvmfsManager struct {
	CvmfsRepo string
}

func (cm CvmfsManager) StartTransaction() error {
	cmd := exec.Command("cvmfs_server", "transaction", publisherConfig.CvmfsRepo)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("ERROR: failed to open transaction!")
		fmt.Println(err)
		fmt.Println(out)
		return err
	}
	fmt.Println("Started transaction...")
	return nil
}

func (cm CvmfsManager) ImportTarball(src, digest string) error {
	dst := path.Join("/cvmfs", cm.CvmfsRepo, "layers", digest)

	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		fmt.Println("Failed to create destination!")
		fmt.Println(err)
		return err
	}
	tarCmd := fmt.Sprintf("tar xf %s -C %s", src, dst)

	cmd := exec.Command("bash", "-c", tarCmd)

	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("ERROR: failed to extract!")
		fmt.Printf("Command was: %s\n", tarCmd)
		fmt.Println(err)
		fmt.Println(string(out))
		return err
	}
	fmt.Println("Extracted")
	return nil
}

func (cm CvmfsManager) PublishTransaction() error {
	cmd := exec.Command("cvmfs_server", "publish")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("ERROR: failed to publish!")
		fmt.Println(err)
		fmt.Println(string(out))
		return err
	} else {
		fmt.Println("Published transaction!")
		fmt.Println(string(out))
		return nil
	}
}

func decodePayload(r *http.Request) (obj Object, err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	var payload WebhookPayload
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		fmt.Println(buf.String())
		return obj, err
	}

	obj = payload.Object()
	return obj, nil
}

func publishHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got a request!")

	var obj Object
	var err error
	if obj, err = decodePayload(r); err != nil {
		fmt.Println("Failed to parse request.")
		return
	}

	fmt.Printf("Bucket:\t%s\n", obj.Bucket)
	fmt.Printf("Key:\t%s\n", obj.Key)

	cm := CvmfsManager{CvmfsRepo: publisherConfig.CvmfsRepo}
	if err := cm.StartTransaction(); err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	filepath := path.Join(publisherConfig.MinioStoragePath, obj.Bucket, obj.Key)
	if err := cm.ImportTarball(filepath, obj.Key); err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
	}

	if err := cm.PublishTransaction(); err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Specify path to config file as first argument.")
		os.Exit(-1)
	}

	var err error
	if publisherConfig, err = LoadConfig(os.Args[1]); err != nil {
		fmt.Println("Invalid config.")
		fmt.Println(err)
		os.Exit(-2)
	}

	fmt.Println("Config finished!")

	http.HandleFunc("/", publishHandler)
	http.ListenAndServe("0.0.0.0:3000", nil)

}
