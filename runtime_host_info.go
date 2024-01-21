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

func GetHostInstanceInfo() RutimeHostInfo {
	retval := RutimeHostInfo{
		BuildId: os.Getenv("BUILD_ID"),
		ImageId: os.Getenv("IMAGE_DIGEST"),
	}

	req, err := http.NewRequest("GET", "http://metadata.google.internal/instance/id", nil)
	if err != nil {
		log.Printf("GetHostInfo: Error metadata request: %s", err.Error())
		return retval
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("GetHostInfo: Error metadata get: %s", err.Error())
		return retval
	}
	defer resp.Body.Close()

	instanceID, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetHostInfo: Error metadata read: %s", err.Error())
	}
	retval.InstanceId = string(instanceID)
	return retval
}
