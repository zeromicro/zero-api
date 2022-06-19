package parser

import (
	"testing"

	"github.com/zeromicro/zero-api/token"
)

var validFiles = []string{
	"../testdata/demo.api",
}

func TestParse(t *testing.T) {
	for _, name := range validFiles {
		_, err := ParseFile(token.NewFileSet(), name, nil, AllErrors)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", name, err)
		}
	}
}
