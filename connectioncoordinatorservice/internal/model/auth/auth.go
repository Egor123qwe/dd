package auth

type TokenType string

const (
	BasicTokenType = "Basic"
	NoAppTokenType = "NoApp"
)

type Token struct {
	Value string
	Type  TokenType
}
