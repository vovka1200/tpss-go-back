package users

import "time"

type User struct {
	Id       string    `json:"id"`
	Username string    `json:"username"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`
	Groups   []string  `json:"groups"`
}
