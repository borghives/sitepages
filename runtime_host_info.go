package sitepages

import (
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RutimeHostInfo struct {
	Id        primitive.ObjectID `bson:"_id"`
	BuildId   string             `bson:"build_id"`
	ImageId   string             `bson:"image_id"`
	AppName   string             `bson:"app_name"`
	StartTime time.Time          `bson:"start_time"`
	EndTime   time.Time          `bson:"end_time"`
}

func NewHostInstanceInfo() RutimeHostInfo {
	return RutimeHostInfo{
		Id:        primitive.NewObjectID(),
		BuildId:   os.Getenv("BUILD_ID"),
		ImageId:   os.Getenv("IMAGE_DIGEST"),
		AppName:   strings.Join(os.Args, " "),
		StartTime: time.Now(),
	}
}
