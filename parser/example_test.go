package parser_test

import (
	"fmt"

	"github.com/zeromicro/zero-api/parser"
	"github.com/zeromicro/zero-api/token"
)

func ExampleParseFile() {
	fset := token.NewFileSet()

	src := `// api语法版本
syntax = "v1"
`
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(f.SyntaxDecl.SyntaxName.Value)

	// output:
	// "v1"
}
