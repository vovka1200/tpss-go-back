package client

import "time"

type Client struct {
	Id      string    `json:"id"`
	Created time.Time `json:"created"`
	Name    string    `json:"name"`
}
