package eokvin // import "github.com/veonik/eokvin"

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ShortURL struct {
	url.URL

	Original *url.URL
}

type Client struct {
	http *http.Client

	Endpoint string
	Token string
}

func NewClient(token string) *Client {
	return &Client{
		http: &http.Client{},
		Endpoint: "https://eok.vin/new",
		Token: token,
	}
}

func NewInsecureClient(token string) *Client {
	return &Client{
		http: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
		Endpoint: "https://localhost:3000/new",
		Token: token,
	}
}

func (c *Client) NewShortURLString(lu string) (*ShortURL, error) {
	ou, err := url.Parse(lu)
	if err != nil {
		return nil, err
	}
	return c.NewShortURL(ou)
}

func (c *Client) NewShortURL(ou *url.URL) (*ShortURL, error) {
	resp, err := c.http.PostForm(
		c.Endpoint,
		url.Values{"token": {c.Token}, "url": {ou.String()}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := struct {
		Error string `json:"error"`
		ShortURL string `json:"short-url"`
	}{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	u, err := url.Parse(r.ShortURL)
	if err != nil {
		return nil, err
	}
	return &ShortURL{URL: *u, Original: ou}, nil
}
