package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	"time"

	"github.com/veonik/eokvin"
)

var token string
var endpoint string
var insecure bool
var ttl time.Duration

func main() {
	flag.StringVar(&token, "token", "", "Secret token, required")
	flag.StringVar(&endpoint, "endpoint", "https://localhost:3000/new", "URL for the running eokvin server")
	flag.BoolVar(&insecure, "insecure", false,"If enabled, allows endpoint to be insecure")
	flag.DurationVar(&ttl, "ttl", 12 * time.Hour, "Short URL expires after this long")
	flag.Parse()

	if token == "" {
		log.Fatal("token cannot be blank")
	}
	if endpoint == "" {
		log.Fatal("endpoint cannot be blank")
	}

	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("expected 1 argument, not %d", len(os.Args))
	}

	u := args[0]

	var c *eokvin.Client
	if !insecure {
		c = eokvin.NewClient(endpoint, token)
	} else {
		c = eokvin.NewInsecureClient(endpoint, token)
	}
	s, err := c.NewShortURLString(u, ttl)
	if err != nil {
		log.Fatalf("error creating short URL: %s", err.Error())
	}
	fmt.Printf("Short URL: %s\t(valid until %s)\n", s, time.Now().Add(ttl).Format("Jan 02 15:04 MST"))
}
