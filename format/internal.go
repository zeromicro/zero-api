package format

import (
	"bytes"

	"github.com/zeromicro/zero-api/printer"

	"github.com/zeromicro/zero-api/ast"
	"github.com/zeromicro/zero-api/parser"
	"github.com/zeromicro/zero-api/token"
)

func parse(fset *token.FileSet, filename string, src []byte) (file *ast.File, err error) {
	return parser.ParseFile(fset, filename, src, parseMode)
}

func format(
	fset *token.FileSet,
	file *ast.File,
	cfg printer.Config) ([]byte, error) {
	var buf bytes.Buffer
	err := cfg.Fprint(&buf, fset, file)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
