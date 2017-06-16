package main

import (
	"fmt"
	"path"
)

func publishWorker() {
	for {
		fmt.Println("Waiting for new obj to be published...")
		obj := <-publisherChannel
		controlChannel <- obj.Key
		fmt.Println("New iteration")

		if err := cm.StartTransaction(); err != nil {
			fmt.Println(err)
			cm.AbortTransaction()
			break
		}

		filepath := path.Join(publisherConfig.MinioStoragePath, obj.Bucket, obj.Key)
		if err := cm.ImportTarball(filepath, obj.Key); err != nil {
			fmt.Println(err)
			cm.AbortTransaction()
			break
		}

		if err := cm.PublishTransaction(); err != nil {
			fmt.Println(err)
			cm.AbortTransaction()
			break
		}
		controlChannel <- ""
	}
}

func statusWorker() {
	var m string

	for {
		select {
		case key := <-controlChannel:
			fmt.Println("got status update")
			m = key
		case req := <-statusChannel:
			fmt.Println("got status request")
			if m == req.Key {
				fmt.Println("publishing")
				req.Status = "publishing"
			} else if cm.LookupLayer(req.Key) {
				fmt.Println("done")
				req.Status = "done"
			} else {
				fmt.Println("unknown")
				req.Status = "unknown"
			}
			statusChannel <- req
		}
	}

}
