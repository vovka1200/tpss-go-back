package group

import "time"

type Group struct {
	Id      string    `json:"id"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Members []string  `json:"members"`
}
