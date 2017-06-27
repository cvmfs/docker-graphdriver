package main

import (
	"fmt"
	"path"
)

func publishWorker() {
	for {
		obj := <-publisherChannel
		controlChannel <- obj.Key
		fmt.Println("Publishing started!")

		if err := cm.StartTransaction(); err != nil {
			fmt.Println(err)
			cm.AbortTransaction()
			controlChannel <- ""
			continue
		}

		filepath := path.Join(publisherConfig.MinioStoragePath, obj.Bucket, obj.Key)
		if err := cm.ImportTarball(filepath, obj.Key); err != nil {
			fmt.Println(err)
			cm.AbortTransaction()
			controlChannel <- ""
			continue
		}

		if err := cm.PublishTransaction(); err != nil {
			fmt.Println(err)
			cm.AbortTransaction()
			controlChannel <- ""
			continue
		}
		controlChannel <- ""
		fmt.Println("Publishing finished!")
	}
}

func statusWorker() {
	var m string

	for {
		select {
		case key := <-controlChannel:
			m = key
		case req := <-statusChannel:
			if m == req.Key {
				req.Status = "publishing"
			} else if cm.LookupLayer(req.Key) {
				req.Status = "done"
			} else {
				req.Status = "unknown"
			}
			statusChannel <- req
		}
	}

}
