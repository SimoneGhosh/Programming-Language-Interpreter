// Defining Tokens - data structures

package token

// Defining as string = many diff value = know different types of token
// Easy to debug
type TokenType string

// Define tokenTypes = limit # of token type
const (
	ILLEGAL = "ILLEGAL" //token/character we don’t know about
	EOF     = "EOF"     // end of file

	// Identifiers + literals
	IDENT = "IDENT" //add, x, y
	INT   = "INT"   //12345

	// operators
	ASSIGN = "="
	PLUS   = "+"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"fn":  FUNCTION,
	"let": LET,
}

// checks the keywords table to see whether the given identifier is in fact a keyword.
// If it is, it returns the keyword’s TokenType constant. If it isn’t, we just get back token.IDENT,
// which is the TokenType for all user-defined identifiers.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
