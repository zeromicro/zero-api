package token

import (
	"go/token"
	"strconv"
)

// Token is the set of lexical tokens of the zero-api programming language.
type Token token.Token

// The list of tokens.
const (
	// Special tokens

	ILLEGAL Token = iota
	EOF
	COMMENT

	literal_beg
	IDENT

	STRING
	VALUE_STRING
	literal_end

	operator_beg
	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	operator_end

	additional_beg
	// additional tokens, handled in an ad-hoc manner

	TILDE
	additional_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:          "EOF",
	COMMENT:      "COMMENT",
	IDENT:        "IDENT",
	STRING:       "STRING",
	VALUE_STRING: "VALUE_STRING",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	TILDE: "~",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
//
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
//
func (tok Token) IsOperator() bool {
	return (operator_beg < tok && tok < operator_end) || tok == TILDE
}
