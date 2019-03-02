package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

// Runtime configuration values.
var listenHost string
var listenPortHTTPS int
var listenPortHTTP int
var canonicalHost string
var tokenSHA256 string

var tlsKeyFile string
var tlsCertFile string

var rawToken string

// urlTTL contains the time.Duration describing how long a short URL shall
// remain accessible.
const urlTTL = 10 * time.Minute
// urlStore contains URL values that map to short, random string keys.
var urlStore = &store{entries: make(map[itemID]item), ttl: urlTTL}

func init() {
	flag.StringVar(&listenHost, "host", "eok.vin", "Listen hostname")
	flag.IntVar(&listenPortHTTPS, "port", 443, "HTTPS listen port")
	flag.IntVar(&listenPortHTTP, "http-port", 80, "HTTP listen port")
	flag.StringVar(&tlsKeyFile, "key-file", "", "TLS private key file, blank for autocert")
	flag.StringVar(&tlsCertFile, "cert-file", "", "TLS certificate chain file, blank for autocert")
	flag.StringVar(&tokenSHA256, "token", "", "SHA256 of the secret token, used to authenticate")
	flag.StringVar(&rawToken, "hash-token", "", "If given, the sha256 of the given value will be printed")
}

func validate() {
	flag.Parse()

	if len(rawToken) > 0 {
		fmt.Printf("%x\n", sha256.Sum256([]byte(rawToken)))
		os.Exit(0)
		return
	}

	if len(listenHost) == 0 {
		log.Fatal("host cannot be blank")
	}
	if listenPortHTTPS <= 0 {
		log.Fatal("port must be > 0")
	}
	if listenPortHTTP <= 0 {
		log.Fatal("http-port must be > 0")
	}
	if len(tokenSHA256) != 64 {
		log.Fatal("token must be a valid sha256 sum")
	}

	canonicalHost = fmt.Sprintf("https://%s:%d/", listenHost, listenPortHTTPS)
}

func main() {
	validate()

	// Launch the HTTP redirect server.
	go func() {
		if err := listenAndServe(); err != nil {
			log.Println(err.Error())
		}
	}()
	// Launch the HTTPS main server.
	go func() {
		if err := listenAndServeTLS(); err != nil {
			log.Println(err.Error())
		}
	}()
	// Launch the expired entry reaper.
	go func() {
		if err := urlStore.expiredItemReaper(); err != nil {
			log.Println(err.Error())
		}
	}()
	// Yield forever.
	select {}
}

