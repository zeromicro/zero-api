package parser

import (
	"github.com/zeromicro/zero-api/ast"
	"github.com/zeromicro/zero-api/scanner"
	"github.com/zeromicro/zero-api/token"
)

type (
	parser struct {
		file    *token.File
		errors  scanner.ErrorList
		scanner scanner.Scanner

		mode   Mode //
		trace  bool // == (mode&Trace != 0)
		indent int  // indentation used for tracing output

		comments    []*ast.CommentGroup
		leadComment *ast.CommentGroup
		lineComment *ast.CommentGroup

		pos token.Pos
		tok token.Token
		lit string
	}
)

func (p *parser) init(fset *token.FileSet, filename string, src []byte, mode Mode) {
	p.file = fset.AddFile(filename, -1, len(src))
	var m scanner.Mode
	if mode&ParseComments != 0 {
		m = scanner.ScanComments
	}

	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner.Init(p.file, src, eh, m)

	p.mode = mode
	p.trace = mode&Trace != 0
	p.next()
}

// ----------------------------------------------------------------------------
// next

func (p *parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	prev := p.pos
	p.next0()

	if p.tok == token.COMMENT {
		var comment *ast.CommentGroup
		var endline int

		if p.file.Line(p.pos) == p.file.Line(prev) {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup(0)
			if p.file.Line(p.pos) != endline || p.tok == token.EOF {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == token.COMMENT {
			comment, endline = p.consumeCommentGroup(1)
		}

		if endline+1 == p.file.Line(p.pos) {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
}

func (p *parser) next0() {
	if p.trace && p.pos.IsValid() {
		s := p.tok.String()
		switch {
		case p.tok.IsLiteral():
			p.printTrace(s, p.lit)
		default:
			p.printTrace(s)
		}
	}

	p.pos, p.tok, p.lit = p.scanner.Scan()
}

// ----------------------------------------------------------------------------
// comment

func (p *parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	comments = &ast.CommentGroup{
		List: list,
	}
	p.comments = append(p.comments, comments)
	return
}

func (p *parser) consumeComment() (*ast.Comment, int) {
	endline := p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}
	comment := &ast.Comment{
		Slash: p.pos,
		Text:  p.lit,
	}
	p.next0()
	return comment, endline
}

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next()
	return pos
}

func (p *parser) expectSemi() {
	if p.tok != token.RPAREN && p.tok != token.RBRACE {
		switch p.tok {
		case token.COMMA:
			p.errorExpected(p.pos, `";"`)
			fallthrough
		case token.SEMICOLON:
			p.next()
		default:
			p.errorExpected(p.pos, `";"`)
			p.advance()
		}
	}
}

func (p *parser) advance() {
	for ; p.tok != token.EOF; p.next() {
		if p.tok == token.IDENT {
			if p.lit == "import" || p.lit == "type" || p.lit == "info" || p.lit == "service" || p.lit == "@server" {
				return
			}
		}
	}
}

// ----------------------------------------------------------------------------
// parse

func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace(p, "File"))
	}

	if p.errors.Len() != 0 {
		return nil
	}

	doc := p.leadComment
	var syntax *ast.SyntaxDecl
	if p.tok == token.IDENT && ast.SYNTAX.Is(p.lit) {
		syntax = p.parseSyntaxDecl()
	} else {
		// TODO: default syntax
	}

	var imports []*ast.GenDecl
	var info *ast.InfoDecl
	for p.tok == token.IDENT && (ast.IMPORT.Is(p.lit) || ast.INFO.Is(p.lit)) {
		if ast.INFO.Is(p.lit) {
			info = p.parseInfoDecl()
		} else if ast.IMPORT.Is(p.lit) {
			imports = append(imports, p.parseGenDecl(ast.IMPORT, p.parseImportSpec)) // parse import
		}
	}

	// type or service
	var decls []ast.Decl
	for p.tok != token.EOF {
		decls = append(decls, p.parseDecl())
	}

	f := &ast.File{
		Doc:         doc,
		SyntaxDecl:  syntax,
		ImportDecls: imports,
		InfoDecl:    info,
		Decls:       decls,
	}

	// TODO: resolveFile

	return f
}
