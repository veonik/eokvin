package main

import (
	"flag"
	"fmt"
	"os"
	"log"

	"github.com/veonik/eokvin"
)

var token string
var endpoint string
var insecure bool

func main() {
	flag.StringVar(&token, "token", "", "Secret token, required")
	flag.StringVar(&endpoint, "endpoint", "https://eok.vin/new", "URL for the running eokvin server")
	flag.BoolVar(&insecure, "insecure", false,"If enabled, allows endpoint to be insecure")
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
		c = eokvin.NewClient(token)
	} else {
		c = eokvin.NewInsecureClient(token)
	}
	c.Endpoint = endpoint
	s, err := c.NewShortURLString(u)
	if err != nil {
		log.Fatalf("error creating short URL: %s", err.Error())
	}
	fmt.Println(s.String())
}
