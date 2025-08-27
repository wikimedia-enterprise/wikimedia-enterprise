package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/confluentinc/confluent-kafka-go/kafka"

	"wikimedia-enterprise/services/snapshots/config/env"
	"wikimedia-enterprise/services/snapshots/submodules/schema"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const (
	createTopics   = true
	mockTopic      = true
	createS3Bucket = true
	mockS3         = true
	removeMirrorS3 = false
	mirrorS3       = false
)

var kafkaTopics = []*schema.MockTopic{
	{
		Topic:  "aws.structured-data.enwiki-articles-compacted.v1",
		Config: schema.ConfigArticle,
		Type:   schema.Article{},
	},
	{
		Topic:  "aws.structured-contents.enwiki-articles-compacted.v1",
		Config: schema.ConfigAvroStructured,
		Type:   schema.AvroStructured{},
	},
	{
		Topic:  "aws.structured-data.eswiki-articles-compacted.v1",
		Config: schema.ConfigArticle,
		Type:   schema.Article{},
	},
}

// number of times to save each json
var repeat = 5000

const worker = 50

type lineTask struct {
	Line []byte
	Name string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ctx := context.Background()
	en, err := env.New()
	if err != nil {
		log.Panic(err)
	}

	if createTopics {
		createKafkaTopics(ctx, en)
	}

	if mockTopic {
		addEvents(ctx)
	}

	s3api := newS3(en)

	if createS3Bucket {
		doCreateS3Bucket(en, s3api)
	}

	if removeMirrorS3 {
		removeMirror(ctx, en, s3api)
	}

	if mockS3 {
		addObjects(ctx, en, s3api)
	}

	if mirrorS3 {
		mirror(ctx, en, s3api)
	}
}

func createKafkaTopics(ctx context.Context, en *env.Environment) {
	broker := en.KafkaBootstrapServers
	partitions := 10

	topicSpecs := []kafka.TopicSpecification{}
	for _, tpc := range kafkaTopics {

		topicSpec := kafka.TopicSpecification{
			Topic:             tpc.Topic,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
			Config: map[string]string{
				"cleanup.policy": "compact",
			},
		}
		topicSpecs = append(topicSpecs, topicSpec)
	}

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		log.Fatalf("Failed to create admin client: %v\n", err)
	}
	defer adminClient.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	results, err := adminClient.CreateTopics(ctx, topicSpecs)
	if err != nil {
		log.Fatalf("Failed to create topic: %v\n", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			log.Fatalf("Failed to create topic: %v\n", result.Error)
		}
		fmt.Printf("Topic %s created successfully!\n", result.Topic)
	}
}

func addEvents(ctx context.Context) {
	mck, err := schema.NewMock()

	if err != nil {
		log.Panic(err)
	}

	_, fnm, _, _ := runtime.Caller(0)

	for _, tpc := range kafkaTopics {
		fle, err := os.Open(fmt.Sprintf("%s/topics/%s.ndjson", filepath.Dir(fnm), tpc.Topic))

		if err != nil {
			log.Panic(err)
		}

		defer func() {
			if err := fle.Close(); err != nil {
				log.Println(err)
			}
		}()

		tpc.Reader = fle
	}

	if err := mck.Run(ctx, kafkaTopics...); err != nil {
		log.Panic(err)
	}
}

func doCreateS3Bucket(en *env.Environment, s3api s3iface.S3API) {
	_, err := s3api.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(en.AWSBucket),
	})
	if err == nil {
		log.Println("Bucket created successfully.")
	} else {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
			// Already exists
		} else {
			log.Fatalln("Failed to create bucket: ", err)
		}
	}
}

func addObjects(ctx context.Context, en *env.Environment, s3api s3iface.S3API) {

	tasks := make(chan lineTask, 1000)
	var wg sync.WaitGroup

	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				for i := 1; i <= repeat; i++ {
					fn := fmt.Sprintf("%s_%d.json", strings.ReplaceAll(task.Name, " ", "_"), i)
					for _, key := range []string{
						fmt.Sprintf("commons/pages/%s.json", fn),
						fmt.Sprintf("commons/batches/%s/%s.json", time.Now().UTC().Format("2006-01-02"), fn),
					} {
						pin := &s3.PutObjectInput{
							Bucket:             aws.String(en.AWSBucket),
							Key:                aws.String(key),
							Body:               bytes.NewReader(task.Line),
							ContentType:        aws.String("application/json"),
							ContentLength:      aws.Int64(int64(len(task.Line))),
							ContentDisposition: aws.String("attachment"),
						}

						_, err := s3api.PutObjectWithContext(ctx, pin)

						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}()
	}

	_, fnm, _, _ := runtime.Caller(0)
	f, err := os.Open(fmt.Sprintf("%s/s3/commonswiki_namespace_6.ndjson", filepath.Dir(fnm)))

	if err != nil {
		log.Panicf("failed to open input file: %v", err)
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()

		var commons schema.Commons
		if err := json.Unmarshal(line, &commons); err != nil {
			log.Printf("skipping line %d: unable to unmarshal json: %v", lineNumber, err)
			continue
		}

		if commons.Name == "" {
			log.Printf("skipping line %d: missing name field", lineNumber)
			continue
		}

		tasks <- lineTask{
			Line: line,
			Name: commons.Name,
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading input file: %v", err)
	}

	close(tasks)
	wg.Wait()
	log.Println("put all files")
}

func newS3(env *env.Environment) s3iface.S3API {
	cfg := &aws.Config{
		Region:                        aws.String(env.AWSRegion),
		S3DisableContentMD5Validation: aws.Bool(true),
	}

	if len(env.AWSID) > 0 && len(env.AWSKey) > 0 {
		cfg.Credentials = credentials.NewStaticCredentials(env.AWSID, env.AWSKey, "")
	}

	if len(env.AWSURL) > 0 {
		cfg.Endpoint = aws.String(env.AWSURL)
	}

	if strings.HasPrefix(env.AWSURL, "http://") {
		cfg.DisableSSL = aws.Bool(true)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	return s3.New(session.Must(session.NewSession(cfg)))
}

func mirror(ctx context.Context, en *env.Environment, s3api s3iface.S3API) {
	rootDir := "./mock/s3/mirror"

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		return uploadFile(ctx, en, s3api, path, relPath)
	})

	if err != nil {
		log.Fatalf("Error mirroring S3 files: %v", err)
	}

	log.Println("All mirror files uploaded to S3")
}

func uploadFile(ctx context.Context, en *env.Environment, s3api s3iface.S3API, localPath, s3Key string) error {
	file, err := os.Open(localPath) // #nosec G304 - Local path is validated upstream, no user input
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", localPath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %v", localPath, err)
	}

	input := &s3.PutObjectInput{
		Bucket:             aws.String(en.AWSBucket),
		Key:                aws.String(s3Key),
		Body:               file,
		ContentType:        aws.String("application/octet-stream"),
		ContentLength:      aws.Int64(stat.Size()),
		ContentDisposition: aws.String("attachment"),
	}

	_, err = s3api.PutObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload %s to S3: %v", s3Key, err)
	}

	log.Printf("Uploaded: %s -> s3://%s/%s", localPath, en.AWSBucket, s3Key)
	return nil
}

func removeMirror(ctx context.Context, en *env.Environment, s3api s3iface.S3API) {
	rootDir := "./mock/s3/mirror"

	var toDeleteObjects []*s3.ObjectIdentifier

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		toDeleteObjects = append(toDeleteObjects, &s3.ObjectIdentifier{
			Key: aws.String(relPath),
		})

		return nil
	})

	if err != nil {
		log.Fatalf("Error traversing local mirror folder: %v", err)
	}

	if len(toDeleteObjects) == 0 {
		log.Println("No mirror files found locally to delete from S3.")
		return
	}

	// Delete in batches (max 1000 per request)
	const batchSize = 1000
	for i := 0; i < len(toDeleteObjects); i += batchSize {
		end := i + batchSize
		if end > len(toDeleteObjects) {
			end = len(toDeleteObjects)
		}

		batch := toDeleteObjects[i:end]
		output, err := s3api.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(en.AWSBucket),
			Delete: &s3.Delete{
				Objects: batch,
				Quiet:   aws.Bool(false),
			},
		})

		if err != nil {
			log.Fatalf("Failed to delete objects batch: %v", err)
		}

		for _, del := range output.Deleted {
			log.Printf("Deleted: %s", *del.Key)
		}

		for _, errObj := range output.Errors {
			log.Printf("Failed to delete: %s, Error: %s", *errObj.Key, *errObj.Message)
		}
	}

	log.Println("Finished deleting mirror files from S3")
}
