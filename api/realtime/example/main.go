package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Filter struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

type Payload struct {
	Since   time.Time `json:"since"`
	Fields  []string  `json:"fields"`
	Filters []Filter  `json:"filters"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	p := Payload{
		Since:  time.Now().Add(-1 * time.Hour),
		Fields: []string{"name", "identifier", "event.*"},
		Filters: []Filter{
			{Field: "identifier", Value: "Earth"},
		},
	}

	body := bytes.NewBuffer([]byte{})

	if err := json.NewEncoder(body).Encode(&p); err != nil {
		log.Panic(err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:4040/v2/articles", body)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Panic(err)
	}

	if res.StatusCode != http.StatusOK {
		log.Panic(res.Status)
	}

	defer res.Body.Close()
	scn := bufio.NewScanner(res.Body)

	for scn.Scan() {
		fmt.Println(scn.Text())
	}
}
