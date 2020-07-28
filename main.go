package kinestesia

import (
	"context"
	"github.com/nicolasassi/kinestesia/kinesis"
	"github.com/nicolasassi/kinestesia/receivers/pubsub"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	log.Println("starting execution...")
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	client, err := kinesis.NewClient(cctx)
	if err != nil {
		log.Fatal(err)
	}
	s, err := kinesis.NewStreamer(cctx, os.Getenv("STREAM_NAME"), client)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("starting pubsub receiver...")
	pbs, err := pubsub.NewPubSubClient(cctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("selected environmt %s", os.Getenv("ENVIRONMENT"))
	var pbsTopic string
	switch os.Getenv("ENVIRONMENT") {
	case "PRD":
		pbsTopic = os.Getenv("PUBSUB_TOPIC")
	default:
		pbsTopic = os.Getenv("PUBSUB_TOPIC_HML")
	}
	log.Printf("streming data to pusub topic: %s", pbsTopic)
	pbs.AddTopics(pbsTopic)
	log.Printf("streaming data from %s",	os.Getenv("STREAM_NAME"))
	if err := s.Stream(cctx, pbs); err != nil {
		log.Fatal(err)
	}
}