package main

import "net/http"
import "fmt"
import "os"

import "github.com/gorilla/mux"

var cm CvmfsManager

var publisherConfig PublisherConfig
var publisherChannel = make(chan Object, 1)
var controlChannel = make(chan string, 1)
var statusChannel = make(chan StatusRequest, 1)

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

	cm = CvmfsManager{CvmfsRepo: publisherConfig.CvmfsRepo}

	go publishWorker()
	go statusWorker()

	m := mux.NewRouter()
	m.HandleFunc("/", webhookHandler)
	m.HandleFunc("/status/{id}", statusHandler)
	http.Handle("/", m)
	http.ListenAndServe(":3000", nil)
}
