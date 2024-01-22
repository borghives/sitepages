package sitepages

import (
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RutimeHostInfo struct {
	Id      primitive.ObjectID `bson:"_id"`
	BuildId string             `bson:"build_id"`
	ImageId string             `bson:"image_id"`
}

func NewHostInstanceInfo() RutimeHostInfo {
	return RutimeHostInfo{
		Id:      primitive.NewObjectID(),
		BuildId: os.Getenv("BUILD_ID"),
		ImageId: os.Getenv("IMAGE_DIGEST"),
	}
}
