package format

import (
	"github.com/zeromicro/zero-api/parser"
	"github.com/zeromicro/zero-api/printer"
	"github.com/zeromicro/zero-api/token"
)

var config = printer.Config{
	Mode:     printer.UseSpaces | printer.TabIndent,
	Tabwidth: 8,
}

const parseMode = parser.ParseComments

func Source(src []byte, filename ...string) ([]byte, error) {
	var fname string
	if len(filename) > 0 {
		fname = filename[0]
	}

	fset := token.NewFileSet()
	file, err := parse(fset, fname, src)
	if err != nil {
		return nil, err
	}
	return format(fset, file, config)
}
