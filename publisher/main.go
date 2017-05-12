package main

import "net/http"
import "fmt"
import "encoding/json"
import "bytes"
import "os/exec"
import "path"
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

func publishHandler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	var payload WebhookPayload
	json.Unmarshal(buf.Bytes(), &payload)

	obj := payload.Object()

	fmt.Printf("Bucket:\t%s\n", obj.Bucket)
	fmt.Printf("Key:\t%s\n", obj.Key)

	// start transaction
	var out bytes.Buffer
	cmd_transaction := exec.Command("cvmfs_server", "transaction")
	cmd_transaction.Stdout = &out
	err := cmd_transaction.Run()
	if err != nil {
		fmt.Println("ERROR: failed to open transaction!")
	}
	fmt.Println("Started transaction...")
	fmt.Println(out.String())

	// unpack the layer
	repo := "docker2cvmfs-ci.cern.ch"
	src := path.Join("/home/ubuntu/minio/export/", obj.Bucket, obj.Key)
	dst := path.Join("/cvmfs", repo, "layers", obj.Key)

	os.Mkdir(dst, os.ModePerm)
	tarCmd := fmt.Sprintf("tar xf %s -C %s", src, dst)

	fmt.Printf("Command is: %s\n", tarCmd)

	cmd_extract := exec.Command("bash", "-c", tarCmd)
	cmd_extract.Stdout = &out
	err = cmd_extract.Run()
	if err != nil {
		fmt.Println("ERROR: failed to extract!")
	}
	fmt.Println("Extracted")
	fmt.Println(out.String())

	cmd_publish := exec.Command("cvmfs_server", "publish")
	cmd_publish.Stdout = &out
	err = cmd_publish.Run()
	if err != nil {
		fmt.Println("ERROR: failed to publish!")
	}
	fmt.Println("Published transaction!")
	fmt.Println(out.String())
}

func main() {
	http.HandleFunc("/", publishHandler)
	http.ListenAndServe("0.0.0.0:3000", nil)

}
