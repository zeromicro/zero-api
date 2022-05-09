package scanner

import (
	"fmt"
	"go/scanner"
	"io"
)

type (
	Error = scanner.Error

	ErrorList = scanner.ErrorList
)

func (s *Scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(s.file.Pos(offs)), msg)
	}
	s.ErrorCount++
}

func (s *Scanner) errorf(offs int, format string, args ...interface{}) {
	s.error(offs, fmt.Sprintf(format, args...))
}

// PrintError is a utility function that prints a list of errors to w,
// one error per line, if the err parameter is an ErrorList. Otherwise
// it prints the err string.
//
func PrintError(w io.Writer, err error) {
	if list, ok := err.(ErrorList); ok {
		for _, e := range list {
			_, _ = fmt.Fprintf(w, "%s\n", e)
		}
	} else if err != nil {
		_, _ = fmt.Fprintf(w, "%s\n", err)
	}
}
