// Command eokvin is a private, single-user, self-hosted URL shortening server.
//
// Short URLs expire after a time and are created by way of HTTP request
// with a static secret token required for authentication. Upon creation, the
// requester will receive the newly generated short URL which can be freely
// accessed by anyone with the link until it expires.
package main

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Runtime configuration values.
var listenHost string
var listenPortHTTPS int
var listenPortHTTP int
var canonicalHost string
var tokenSHA256 string
var urlTTL time.Duration

var tlsKeyFile string
var tlsCertFile string

var rawToken string

// DefaultTTL the duration a short URL shall remain accessible.
const DefaultTTL = 60 * time.Minute

// urlStore contains URL values that map to short, random string keys.
var urlStore *store

func init() {
	flag.StringVar(&listenHost, "host", "localhost", "Listen hostname")
	flag.IntVar(&listenPortHTTPS, "port", 443, "HTTPS listen port")
	flag.IntVar(&listenPortHTTP, "http-port", 80, "HTTP listen port")
	flag.DurationVar(&urlTTL, "url-ttl", DefaultTTL, "Short URLs expire after this delay")
	flag.StringVar(&tlsKeyFile, "key-file", "", "TLS private key file, blank for autocert")
	flag.StringVar(&tlsCertFile, "cert-file", "", "TLS certificate chain file, blank for autocert")
	flag.StringVar(&tokenSHA256, "token", "", "SHA256 of the secret token, used to authenticate")

	flag.StringVar(&rawToken, "hash-token", "", "If given, the sha256 of the value will be printed")
}

func main() {
	// Read and validate configuration.
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	urlStore = &store{entries: make(map[itemID]item), ttl: urlTTL}

	// Launch the HTTP->HTTPS redirect server.
	go func() {
		log.Println("starting HTTP")
		if err := listenAndServeRedirect(); err != nil {
			log.Println(err.Error())
		}
	}()
	// Launch the HTTPS main server.
	go func() {
		log.Println("starting TLS")
		if err := listenAndServeTLS(); err != nil {
			log.Println(err.Error())
		}
	}()
	// Launch the expired entry reaper.
	go func() {
		log.Println("starting expired item reaper")
		if err := urlStore.expiredItemReaper(); err != nil {
			log.Println(err.Error())
		}
	}()
	log.Println("started all goroutines")
	// Yield forever.
	select {}
}

// parseFlags reads command line options and ensures the program state is
// configured properly, returning an error if it is not.
func parseFlags() error {
	flag.Parse()

	if len(rawToken) > 0 {
		fmt.Printf("%x\n", sha256.Sum256([]byte(rawToken)))
		os.Exit(0)
		return nil
	}

	if len(listenHost) == 0 {
		return errors.New("host cannot be blank")
	}
	if listenPortHTTPS <= 0 {
		return errors.New("port must be > 0")
	}
	if listenPortHTTP <= 0 {
		return errors.New("http-port must be > 0")
	}
	if len(tokenSHA256) != sha256.BlockSize {
		return errors.New("token must be a valid sha256 sum")
	}

	parts := []string{"https://", listenHost}
	if listenPortHTTPS != 443 {
		parts = append(parts, ":", strconv.Itoa(listenPortHTTPS))
	}

	canonicalHost = fmt.Sprintf("%s", strings.Join(parts, ""))

	log.Println("listen host:", listenHost)
	log.Println("listen tls port:", listenPortHTTPS)
	log.Println("listen http port:", listenPortHTTP)
	log.Println("canonical host:", canonicalHost)
	log.Println("url ttl:", urlTTL)
	log.Println("token:", tokenSHA256)

	return nil
}
