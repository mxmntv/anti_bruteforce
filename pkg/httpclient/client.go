package httpclient

import (
	"net/http"
	"time"
)

type Client struct {
	client *http.Client
}

func NewDefaultClient() *Client {
	client := &Client{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    100,
				IdleConnTimeout: 60 * time.Second,
			},
		},
	}
	return client
}
