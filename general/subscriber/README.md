# Subscriber

This is a package to provide graceful shutdown for kafka handlers. Example follows:

```go

cnt, err := container.New()

if err != nil {
  log.Panic(err)
}

app := func(env *env.Environment, sbs subscriber.Subscriber, prs handler.Parameters) error {
  ctx := context.Background()
  ers := make(chan error, 1000000)

  go func() {
    for err := range ers {
      log.Println(err)
    }
  }()

  return sbs.Subscribe(
    ctx,
    env.TopicArticles,
    handler.NewExample(&prs),
    ers,
  )
}

if err := cnt.Invoke(app); err != nil {
  log.Panic(err)
}

```
