package scanner

import (
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"github.com/zeromicro/zero-api/token"
)

const (
	ScanComments    Mode = 1 << iota // return comments as COMMENT tokens
	dontInsertSemis                  // do not automatically insert semicolons - for testing only
)

type (
	// An ErrorHandler may be provided to Scanner.Init. If a syntax error is
	// encountered and a handler was installed, the handler is called with a
	// position and an error message. The position points to the beginning of
	// the offending token.
	//
	ErrorHandler func(pos token.Position, msg string)

	// A Scanner holds the scanner's internal state while processing
	// a given text. It can be allocated as part of another data
	// structure but must be initialized via Init before use.
	//
	Scanner struct {
		// immutable state
		file *token.File  // source file handle
		dir  string       // directory portion of file.Name()
		src  []byte       // source
		err  ErrorHandler // error reporting; or nil
		mode Mode         // scanning mode

		// scanning state
		ch         rune // current character
		offset     int  // character offset
		rdOffset   int  // reading offset (position after current character)
		lineOffset int  // current line offset
		insertSemi bool // insert a semicolon before next newline

		// public state - ok to modify
		ErrorCount int // number of errors encountered
	}

	// Mode A mode value is a set of flags (or 0).
	// They control scanner behavior.
	//
	Mode uint
)

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err
	s.mode = mode

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.insertSemi = false
	s.ErrorCount = 0

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}
}

const (
	bom = 0xFEFF // byte order mark, only permitted as very first character
	eof = -1     // end of file
)

func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

func stripCR(b []byte, comment bool) []byte {
	c := make([]byte, len(b))
	i := 0
	for j, ch := range b {
		// In a /*-style comment, don't strip \r from *\r/ (incl.
		// sequences of \r from *\r\r...\r/) since the resulting
		// */ would terminate the comment too early unless the \r
		// is immediately following the opening /* in which case
		// it's ok because /*/ is not closed yet (issue #11151).
		if ch != '\r' || comment && i > len("/*") && c[i-1] == '*' && j+1 < len(b) && b[j+1] == '/' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' && !s.insertSemi || s.ch == '\r' {
		s.next()
	}
}

// Scan scans the next token and returns the token position, the token,
// and its literal string if applicable. The source end is indicated by
// token.EOF.
//
func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
scanAgain:
	s.skipWhitespace()

	pos = s.file.Pos(s.offset)

	if s.ch == -1 {
		if s.insertSemi {
			s.insertSemi = false
			return pos, token.SEMICOLON, "\n"
		}
		tok = token.EOF
		return
	}
	ch := s.ch
	offs := s.offset
	s.next()

	insertSemi := false
	switch ch {
	case '/':
		switch s.ch {
		case '/', '*':
			// comment
			if s.insertSemi && s.findLineEnd() {
				// reset position to the beginning of the comment
				s.ch = '/'
				s.offset = s.file.Offset(pos)
				s.rdOffset = s.offset + 1
				s.insertSemi = false // newline consumed
				return pos, token.SEMICOLON, "\n"
			}
			comment := s.scanComment()
			if s.mode&ScanComments == 0 {
				// skip comment
				s.insertSemi = false // newline consumed
				goto scanAgain
			}
			tok = token.COMMENT
			lit = comment
		default:
			insertSemi = true
			lit = s.scanIdentifier(true, offs)
			tok = token.IDENT
		}
	case '\n':
		s.insertSemi = false
		return pos, token.SEMICOLON, "\n"
	case '"':
		insertSemi = true
		tok = token.STRING
		lit = s.scanString()
	case '`':
		insertSemi = true
		tok = token.STRING
		lit = s.scanRawString()
	case '=':
		tok = token.ASSIGN
	case ':':
		tok = token.COLON
	case ',':
		tok = token.COMMA
	case ';':
		tok = token.SEMICOLON
		lit = ";"
	case '(':
		tok = token.LPAREN
	case ')':
		insertSemi = true
		tok = token.RPAREN
	case '[':
		tok = token.LBRACK
	case ']':
		insertSemi = true
		tok = token.RBRACK
	case '{':
		tok = token.LBRACE
	case '}':
		insertSemi = true
		tok = token.RBRACE
	case '~':
		tok = token.TILDE
	case '*':
		tok = token.MUL
	default:
		insertSemi = true
		lit = s.scanIdentifier(false, offs)
		tok = token.IDENT
	}

	if s.mode&dontInsertSemis == 0 {
		s.insertSemi = insertSemi
	}
	return
}

func (s *Scanner) findLineEnd() bool {
	// initial '/' already consumed

	defer func(offs int) {
		// reset scanner state to where it was upon calling findLineEnd
		s.ch = '/'
		s.offset = offs
		s.rdOffset = offs + 1
		s.next() // consume initial '/' again
	}(s.offset - 1)

	// read ahead until a newline, EOF, or non-comment token is found
	for s.ch == '/' || s.ch == '*' {
		if s.ch == '/' {
			//-style comment always contains a newline
			return true
		}
		/*-style comment: look for newline */
		s.next()
		for s.ch >= 0 {
			ch := s.ch
			if ch == '\n' {
				return true
			}
			s.next()
			if ch == '*' && s.ch == '/' {
				s.next()
				break
			}
		}
		s.skipWhitespace() // s.insertSemi is set
		if s.ch < 0 || s.ch == '\n' {
			return true
		}
		if s.ch != '/' {
			// non-comment token
			return false
		}
		s.next() // consume '/'
	}

	return false
}

func (s *Scanner) scanRawString() string {
	// '`' opening already consumed
	offs := s.offset - 1

	hasCR := false
	for {
		ch := s.ch
		if ch < 0 {
			s.error(offs, "raw string literal not terminated")
			break
		}
		s.next()
		if ch == '`' {
			break
		}
		if ch == '\r' {
			hasCR = true
		}
	}

	lit := s.src[offs:s.offset]
	if hasCR {
		lit = stripCR(lit, false)
	}

	return string(lit)
}
