package entities

import "time"

type Entity struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	Created  time.Time  `json:"created"`
	Updated  time.Time  `json:"updated"`
	Archived *time.Time `json:"archived"`
}
