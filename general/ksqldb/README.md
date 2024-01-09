# Wikimedia Enterprise ksqlDB client

This client provides the bare-bones integration with ksqlDB. It supports both pull and push queries.

## Getting started:

1. Creating new client instance:

   ```go
   client := ksqldb.NewClient("http://locahost:8080")
   ```

1. For basic authentication define the BasicAuth in client constructor:

   ```go
   client := ksqldb.NewClient("http://localhost:8080", func(c *ksqldb.Client) {
       c.HTTPClient = &http.Client{}
       c.BasicAuth = &ksqldb.BasicAuth{Username: os.Getenv("AUTH_USERNAME"), Password: os.Getenv("AUTH_USERNAME")}
   })
   ```

1. Running `pull` query:

   ```go
   client := ksqldb.NewClient("http://localhost:8080", func(c *ksqldb.Client) {
   	c.HTTPClient = &http.Client{}
   })

   hr, rows, err := client.Pull(context.Background(), &ksqldb.QueryRequest{SQL: "SELECT * FROM articles;"})

   for idx, row := range rows {
       // You can access a row value either by calling row.String(idx) (or any other available row method depending on a value type) or rows[idx]
   }
   ```

1. Running `push` query:

   ```go
   client := ksqldb.NewClient("http://localhost:8080")
   cb := func(hr *ksqldb.HeaderRow, row ksqldb.Row) error {
       // You can access a row value by calling row.String(idx) (or any other available row method depending on a value type) or rows[idx]
   }

   err := client.Push(context.Background(), &ksqldb.QueryRequest{SQL: "SELECT * FROM articles EMIT CHANGES;"}, cb)

   if err != nil {
       log.Panic(err)
   }
   ```

1. Using context:

   ```go
   ctx, ctxCancel := context.WithTimeout(context.Background(), 10 * time.Second)
   defer ctxCancel()

   client := ksqldb.NewClient("http://localhost:8080")
   cb := func(hr *ksqldb.HeaderRow, row ksqldb.Row) error {
       // You can access a row value by calling row.String(idx) (or any other available row method depending on a value type) or rows[idx]
   }

   err := client.Push(ctx, &ksqldb.QueryRequest{SQL: "SELECT * FROM articles EMIT CHANGES;"}, cb)

   if err != nil {
       log.Panic(err)
   }
   ```

1. Mapping `ksqldb.Row` to `struct`. This package includes default decoder that will help you map ksqlDB responses to avro schema. Here's an example how to use `Decoder`:

   ```go
   type Article struct {
       TableName  struct{} `json:"-" ksql:"articles"`
       Name       string   `json:"name" avro:"name"`
       Identifier int      `json:"identifier,omitempty" avro:"identifier"`
   }

   client := ksqldb.NewClient("http://localhost:8080")

   cb := func(hr *ksqldb.HeaderRow, row ksqldb.Row) error {
       art := new(Article)

       if err := ksqldb.NewDecoder(row, hr.ColumnNames).Decode(art); err != nil {
           return err
       }

       log.Println(art.Name)

       return nil
   }

   err := client.Push(context.Background(), &ksqldb.QueryRequest{SQL: "SELECT * FROM articles EMIT CHANGES;"}, cb)

   if err != nil {
       log.Panic(err)
   }
   ```
