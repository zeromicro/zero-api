package scanner

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"unicode"
)

// ----------------------------------------------------------------------------
// scanIdentifier
func (s *Scanner) scanIdentifier(quo bool, offs int) string {

	for {
		ch := s.ch
		if quo && ch == ':' { // because isPunctuation contain ':" , wo should parse '/:user'
			quo = s.ch == '/'
			s.next()
			continue
		}

		switch ch := s.ch; {
		case ch < 0,
			isPunctuation(ch),
			ch == '\n' || ch == ' ':
			goto exit
		}

		quo = s.ch == '/'
		s.next()
	}
exit:
	lit := s.src[offs:s.offset]
	return string(lit)
}

// ----------------------------------------------------------------------------

var prefix = []byte("line ")

// scanComment
func (s *Scanner) scanComment() string {
	// initial '/' already consumed; s.ch == '/' || s.ch == '*'
	offs := s.offset - 1 // position of initial '/'
	next := -1           // position immediately following the comment; < 0 means invalid comment
	numCR := 0

	if s.ch == '/' {
		//-style comment
		// (the final '\n' is not considered part of the comment)
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			if s.ch == '\r' {
				numCR++
			}
			s.next()
		}
		// if we are at '\n', the position following the comment is afterwards
		next = s.offset
		if s.ch == '\n' {
			next++
		}
		goto exit
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		if ch == '\r' {
			numCR++
		}
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			next = s.offset
			goto exit
		}
	}

	s.error(offs, "comment not terminated")

exit:
	lit := s.src[offs:s.offset]

	// On Windows, a (//-comment) line may end in "\r\n".
	// Remove the final '\r' before analyzing the text for
	// line directives (matching the compiler). Remove any
	// other '\r' afterwards (matching the pre-existing be-
	// havior of the scanner).
	if numCR > 0 && len(lit) >= 2 && lit[1] == '/' && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}

	// interpret line directives
	// (//line directives must start at the beginning of the current line)
	if next >= 0 /* implies valid comment */ && (lit[1] == '*' || offs == s.lineOffset) && bytes.HasPrefix(lit[2:], prefix) {
		s.updateLineInfo(next, offs, lit)
	}

	if numCR > 0 {
		lit = stripCR(lit, lit[1] == '*')
	}

	return string(lit)
}

// updateLineInfo parses the incoming comment text at offset offs
// as a line directive. If successful, it updates the line info table
// for the position next per the line directive.
func (s *Scanner) updateLineInfo(next, offs int, text []byte) {
	// extract comment text
	if text[1] == '*' {
		text = text[:len(text)-2] // lop off trailing "*/"
	}
	text = text[7:] // lop off leading "//line " or "/*line "
	offs += 7

	i, n, ok := trailingDigits(text)
	if i == 0 {
		return // ignore (not a line directive)
	}
	// i > 0

	if !ok {
		// text has a suffix :xxx but xxx is not a number
		s.error(offs+i, "invalid line number: "+string(text[i:]))
		return
	}

	var line, col int
	i2, n2, ok2 := trailingDigits(text[:i-1])
	if ok2 {
		//line filename:line:col
		i, i2 = i2, i
		line, col = n2, n
		if col == 0 {
			s.error(offs+i2, "invalid column number: "+string(text[i2:]))
			return
		}
		text = text[:i2-1] // lop off ":col"
	} else {
		//line filename:line
		line = n
	}

	if line == 0 {
		s.error(offs+i, "invalid line number: "+string(text[i:]))
		return
	}

	// If we have a column (//line filename:line:col form),
	// an empty filename means to use the previous filename.
	filename := string(text[:i-1]) // lop off ":line", and trim white space
	if filename == "" && ok2 {
		filename = s.file.Position(s.file.Pos(offs)).Filename
	} else if filename != "" {
		// Put a relative filename in the current directory.
		// This is for compatibility with earlier releases.
		// See issue 26671.
		filename = filepath.Clean(filename)
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(s.dir, filename)
		}
	}

	s.file.AddLineColumnInfo(next, filename, line, col)
}

func trailingDigits(text []byte) (int, int, bool) {
	i := bytes.LastIndexByte(text, ':') // look from right (Windows filenames may contain ':')
	if i < 0 {
		return 0, 0, false // no ":"
	}
	// i >= 0
	n, err := strconv.ParseUint(string(text[i+1:]), 10, 0)
	return i + 1, int(n), err == nil
}

// ----------------------------------------------------------------------------
// scanString
func (s *Scanner) scanString() string {
	// '"' opening already consumed
	offs := s.offset - 1

	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			s.scanEscape('"')
		}
	}
	return string(s.src[offs:s.offset])
}

// scanEscape parses an escape sequence where rune is the accepted
// escaped quote. In case of a syntax error, it stops at the offending
// character (without consuming it) and returns false. Otherwise
// it returns true.
func (s *Scanner) scanEscape(quote rune) bool {
	offs := s.offset

	var n int
	var base, max uint32
	switch s.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		s.next()
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		s.next()
		n, base, max = 2, 16, 255
	case 'u':
		s.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		s.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "unknown escape sequence"
		if s.ch < 0 {
			msg = "escape sequence not terminated"
		}
		s.error(offs, msg)
		return false
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(s.ch))
		if d >= base {
			msg := fmt.Sprintf("illegal character %#U in escape sequence", s.ch)
			if s.ch < 0 {
				msg = "escape sequence not terminated"
			}
			s.error(s.offset, msg)
			return false
		}
		x = x*base + d
		s.next()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		s.error(offs, "escape sequence is invalid Unicode code point")
		return false
	}

	return true
}
