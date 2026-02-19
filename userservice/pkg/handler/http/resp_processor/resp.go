package resp_processor

type Response struct {
	Data       any                 `json:"data"`
	Messages   []string            `json:"messages"`
	Validation map[string][]string `json:"validation"`
}
