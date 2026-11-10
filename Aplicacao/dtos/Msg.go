package dtos

type Msg struct {
	Role    string `json:"role"` // "user" ou "assistant"
	Content string `json:"content"`
}