package dtos

type Config struct {
	WorkDir  string `json:"workDir"`
	Port     int    `json:"port"`
	Usuario  string `json:"usuario"`
	Provider string `json:"provider"` // "gemini" ou "openai" (qualquer API compatível)
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseUrl"` // só para provider "openai"
}