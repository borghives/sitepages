package sitepages

import (
	"io"
	"log"
	"net/http"
	"os"
)

type RutimeHostInfo struct {
	InstanceId string `bson:"instance_id"`
	BuildId    string `bson:"build_id"`
	ImageId    string `bson:"image_id"`
}

func GetHostInfo() RutimeHostInfo {
	resp, err := http.Get("http://metadata.google.internal/instance/id")
	if err != nil {
		log.Printf("GetHostInfo: Error metadata get: %s", err.Error())
	}
	defer resp.Body.Close()

	instanceID, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetHostInfo: Error metadata read: %s", err.Error())
	}

	return RutimeHostInfo{
		InstanceId: string(instanceID),
		BuildId:    os.Getenv("BUILD_ID"),
		ImageId:    os.Getenv("IMAGE_DIGEST"),
	}
}
