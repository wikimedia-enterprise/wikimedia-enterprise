package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/event-bridge/config/env"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

func main() {
	env, err := env.New()

	if err != nil {
		log.Panic(err)
	}

	url := fmt.Sprintf("http://%s", env.SchemaRegistryURL)

	for {
		time.Sleep(time.Second * 2)

		// #nosec G107
		res, err := http.Get(url)

		if err != nil {
			log.Println(err)
			continue
		}

		if res.StatusCode == http.StatusOK {
			break
		}

		log.Println(res.Status)
	}

	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
		"group.id":          uuid.NewString(),
		"bootstrap.servers": env.KafkaBootstrapServers,
	})

	if err != nil {
		log.Panic(err)
	}

	defer cons.Close()

	if err := cons.SubscribeTopics([]string{
		"aws.event-bridge.article-delete.v1",
		"aws.event-bridge.article-update.v1",
		"aws.event-bridge.article-visibility-change.v1"}, nil); err != nil {
		log.Panic(err)
	}

	ctx := context.Background()
	shl := schema.NewHelper(schema.NewRegistry(url), nil)

	for {
		msg, err := cons.ReadMessage(time.Minute * 1)

		if err != nil {
			log.Println(err)
			continue
		}

		key := new(schema.Key)

		if err := shl.Unmarshal(ctx, msg.Key, key); err != nil {
			log.Println(err)
			continue
		}

		article := new(schema.Article)

		if err := shl.Unmarshal(ctx, msg.Value, article); err != nil {
			log.Println(err)
			continue
		}

		kData, err := json.Marshal(key)

		if err != nil {
			log.Println(err)
			continue
		}

		vData, err := json.Marshal(article)

		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("topic: %s\n", *msg.TopicPartition.Topic)
		log.Printf("offset: %d\n", msg.TopicPartition.Offset)
		log.Printf("partition: %d\n", msg.TopicPartition.Partition)
		log.Printf("key: %s\n", string(kData))
		log.Printf("value: %s\n", string(vData))
	}
}
