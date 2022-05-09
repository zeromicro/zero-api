package scanner_test

import (
	"fmt"

	"github.com/zeromicro/zero-api/scanner"
	"github.com/zeromicro/zero-api/token"
)

func ExampleScanner_Scan() {
	// src is the input that we want to tokenize.
	src := []byte(`post /foo (Foo) returns (Bar)`)

	// Initialize the scanner.
	var s scanner.Scanner
	fset := token.NewFileSet()                      // positions are relative to fset
	file := fset.AddFile("", fset.Base(), len(src)) // register input "file"
	s.Init(file, src, nil /* no error handler */, scanner.ScanComments)

	// Repeated calls to Scan yield the token sequence found in the input.
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		fmt.Printf("%s\t%s\t%q\n", fset.Position(pos), tok, lit)
	}

	// output:
	// 1:1	IDENT	"post"
	// 1:6	IDENT	" /foo"
	// 1:11	(	""
	// 1:12	IDENT	"Foo"
	// 1:15	)	""
	// 1:17	IDENT	"returns"
	// 1:25	(	""
	// 1:26	IDENT	"Bar"
	// 1:29	)	""
	// 1:30	;	"\n"
}
