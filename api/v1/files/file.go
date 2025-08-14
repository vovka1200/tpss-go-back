package files

type File struct {
	Id   string `json:"id"`
	Data []byte `json:"data"`
}
