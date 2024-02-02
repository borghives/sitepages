package sitepages

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
)

func DailyFunction(ctx context.Context, event *pubsub.Message) error {
	// Your daily logic here
	log.Println("This function is running once a day!")
	return nil
}
