# Wikimedia Enterprise Schema

This repository contains schema evolution for the project and used as a package in services to make sure schema is consistent across the project.

## Schema helper:

1. Running schema synchronization with schema registry (this will current article schema and synchronize it with schema registry):

    ```go
	sch, err := schema.
		NewHelper("http://localhost:8085").
		Sync(context.Background(), "my-topic", schema.ConfigArticle)

	if err != nil {
		log.Panic(err)
	}

	log.Println(sch.ID, sch.AVRO, sch.Schema)

	article := new(schema.Article) // my awesome data that I want to send to kafka.
	data, err := schema.Marshal(sch.ID, sch.AVRO, article)

	if err != nil {
		log.Panic(err)
	}

	log.Println(schema.GetID(data))
    ```

1. Getting schema from registry by id (example how to `unmarshal` the message):

    ```go
	data := []byte{} // my message data from kafka

	id, err := schema.GetID(data)

	if err != nil {
		log.Panic(err)
	}

	sch, err := schema.
		NewHelper("http://localhost:8085").
		Get(context.Background(), id)

	if err != nil {
		log.Panic(err)
	}

	article := new(schema.Article)

	if err := schema.Unmarshal(sch.AVRO, data, article); err != nil {
		log.Panic(err)
	}

	log.Println(article)
    ```

## Schema registry:

1. Using authentication or custom `http` client:

    ```go
	reg := schema.NewRegistry("http://locahost:8080", func(reg *schema.Registry) {
		cl.HTTPClient = new(http.Client)
		cl.BasicAuth = &schema.BasicAuth{
			Username: "user",
			Password: "sql",
		}
	})
    ```

2. Create subject example:

    ```go
	sch, err := schema.
		NewRegistry("http://localhost:8080").
		CreateSubject(context.Background(), "my-subject", &schema.Subject{
			SchemaType: schema.SchemaTypeAVRO,
			Schema:     `{"type":"record","name":"Key","namespace":"wikimedia_enterprise.general.schema","fields":[{"name":"identifier","type":"string"},{"name":"type","type":"string"}]}`,
		})

	if err != nil {
		log.Panic(err)
	}

	log.Println(sch)
    ```
