package matrix

type Matrix []Rule

type Rule struct {
	Object  string   `json:"object"`
	Methods []string `json:"methods"`
}
