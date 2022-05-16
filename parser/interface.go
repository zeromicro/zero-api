package parser

import (
	"errors"
	"io"
	"os"

	"github.com/zeromicro/zero-api/ast"
	"github.com/zeromicro/zero-api/token"
)

type Mode uint

const (
	iotaMode Mode = 1 << iota
	ParseComments
	Trace
	AllErrors
)

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case io.Reader:
			return io.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return os.ReadFile(filename)
}

func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error) {
	if fset == nil {
		panic("parser.ParseFile: not token.FileSet provided (fset == nil)")
	}

	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if e := recover(); err != nil {
			if _, ok := e.(bailout); ok {
				panic(e)
			}
		}

		if f == nil {
			f = &ast.File{}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	p.init(fset, filename, text, mode)
	f = p.parseFile()
	return
}
