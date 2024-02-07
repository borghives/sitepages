package sitepages

import (
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RutimeHostInfo struct {
	Id         primitive.ObjectID `bson:"_id"`
	BuildId    string             `bson:"build_id"`
	ImageId    string             `bson:"image_id"`
	AppName    string             `bson:"app_name"`
	AppCommand string             `bson:"app_command"`
	CreatedAt  time.Time          `bson:"created_at"`
	EnvVars    []string           `bson:"env_vars"`
}

func getNonSecretEnvVars() []string {
	envVars := []string{}
	for _, env := range os.Environ() {
		//exclude secrets
		if !strings.HasPrefix(env, "CSRF_LATEST") &&
			!strings.HasPrefix(env, "SESSION_LATEST") &&
			!strings.HasPrefix(env, "SECRET_") {
			envVars = append(envVars, env)
		}
	}
	return envVars
}
func NewHostInstanceInfo() RutimeHostInfo {

	return RutimeHostInfo{
		Id:         primitive.NewObjectID(),
		BuildId:    os.Getenv("BUILD_ID"),
		ImageId:    os.Getenv("IMAGE_DIGEST"),
		AppName:    os.Getenv("APP_NAME"),
		AppCommand: strings.Join(os.Args, " "),
		EnvVars:    getNonSecretEnvVars(),
		CreatedAt:  time.Now(),
	}
}
