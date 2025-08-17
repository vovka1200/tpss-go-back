package matrix

type Rules []Rule

type Rule struct {
	Object      string   `json:"object"`
	Access      []string `json:"access"`
	Description *string  `json:"description"`
}
