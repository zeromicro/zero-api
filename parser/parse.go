package parser

import (
	"github.com/zeromicro/zero-api/ast"
	"github.com/zeromicro/zero-api/token"
)

func (p *parser) parseSyntaxDecl() *ast.SyntaxDecl {
	if p.trace {
		defer un(trace(p, "Syntax"))
	}

	pos := p.expect(token.IDENT)
	assignPos := p.expect(token.ASSIGN)

	namePos := p.pos
	var name string

	if p.tok == token.STRING {
		name = p.lit
	}
	p.expect(token.STRING)
	p.expectSemi()

	return &ast.SyntaxDecl{
		TokPos:     pos,
		Assign:     assignPos,
		SyntaxName: &ast.BasicLit{ValuePos: namePos, Kind: token.STRING, Value: name},
	}
}

func (p *parser) parseInfoDecl() *ast.InfoDecl {
	if p.trace {
		defer un(trace(p, "Info"))
	}
	pos := p.expect(token.IDENT)
	lparen := p.expect(token.LPAREN)
	elements := p.parseElementList()
	rparen := p.expect(token.RPAREN)
	p.expectSemi()

	return &ast.InfoDecl{
		TokPos:   pos,
		Lparen:   lparen,
		Elements: elements,
		Rparen:   rparen,
	}
}

func (p *parser) parseElementList() []*ast.KeyValueExpr {
	if p.trace {
		defer un(trace(p, "ElementList"))
	}

	var kvs []*ast.KeyValueExpr
	for p.tok != token.RPAREN && p.tok != token.EOF {
		kvs = append(kvs, p.parseElement(true))
		p.expectSemi()
	}
	return kvs
}

func (p *parser) parseElement(expectColon bool) *ast.KeyValueExpr {
	if p.trace {
		defer un(trace(p, "Element"))
	}

	key := p.parseIdent(false)
	var colonPos token.Pos
	if expectColon {
		colonPos = p.expect(token.COLON)
	}
	value := &ast.BasicLit{
		ValuePos: p.pos,
		Kind:     p.tok,
		Value:    p.lit,
	}
	p.next()

	return &ast.KeyValueExpr{
		Key:   key,
		Colon: colonPos,
		Value: value,
	}
}

type parseSpecFunction func(doc *ast.CommentGroup) ast.Spec

func (p *parser) parseGenDecl(key ast.Keyword, f parseSpecFunction) *ast.GenDecl {
	if p.trace {
		defer un(trace(p, "GenDecl("+string(key)+")"))
	}

	doc := p.leadComment
	pos := p.expect(token.IDENT) // import
	var lparen, rparen token.Pos
	var list []ast.Spec
	if p.tok == token.LPAREN {
		lparen = p.pos
		p.next()
		for iota := 0; p.tok != token.RPAREN && p.tok != token.EOF; iota++ {
			list = append(list, f(p.leadComment))
		}
		rparen = p.expect(token.RPAREN)
		p.expectSemi()
	} else {
		list = append(list, f(doc))
	}
	return &ast.GenDecl{
		Doc:    doc,
		TokPos: pos,
		Key:    key,
		Lparen: lparen,
		Specs:  list,
		Rparen: rparen,
	}
}

func (p *parser) parseImportSpec(doc *ast.CommentGroup) ast.Spec {
	if p.trace {
		defer un(trace(p, "Import"))
	}
	pos := p.pos
	var path string
	if p.tok == token.STRING {
		path = p.lit
	}
	p.expect(token.STRING)
	p.expectSemi()

	return &ast.ImportSpec{
		Doc:     doc,
		Path:    &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: path},
		Comment: p.lineComment,
		EndPos:  0,
	}
}

func (p *parser) parseIdent(identifier bool) *ast.Ident {
	if p.trace {
		defer un(trace(p, "Ident"))
	}
	pos := p.pos
	var name string
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
		if identifier && !token.IsIdentifier(name) {
			p.error(pos, "expect Identifier")
		}
	} else {
		name = "_"
		p.expect(token.IDENT) // use expect() error handling
	}

	return &ast.Ident{
		NamePos: pos,
		Name:    name,
	}
}

// ----------------------------------------------------------------------------
// decl

func (p *parser) parseDecl() ast.Decl {
	if p.trace {
		defer un(trace(p, "Decl"))
	}

	if p.tok != token.IDENT {
		pos := p.pos
		p.errorExpected(pos, "expect declaration")
		p.advance()
		return &ast.BadDecl{
			From: pos,
			To:   p.pos,
		}
	}

	switch {
	case ast.TYPE.Is(p.lit):
		return p.parseGenDecl(ast.TYPE, p.parseTypeSpec)
	case ast.SERVICE.Is(p.lit), ast.SERVEREXT.Is(p.lit):
		return p.parseService()
	}

	pos := p.pos
	p.errorExpected(pos, "expect declaration")
	p.advance()
	return &ast.BadDecl{
		From: pos,
		To:   p.pos,
	}
}

// ----------------------------------------------------------------------------
// struct

func (p *parser) parseTypeSpec(doc *ast.CommentGroup) ast.Spec {
	if p.trace {
		defer un(trace(p, "TypeSpec"))
	}

	ident := p.parseIdent(true)
	ty := p.parseStructType() // type spec only support structType
	p.expectSemi()

	return &ast.TypeSpec{
		Doc:     doc,
		Name:    ident,
		Type:    ty,
		Comment: p.lineComment,
	}
}

func (p *parser) parseStructType() *ast.StructType {
	if p.trace {
		defer un(trace(p, "StructType"))
	}

	var structPos, lbrace token.Pos
	if p.tok == token.LBRACE {
		lbrace = p.expect(token.LBRACE)
	} else if p.tok == token.IDENT && p.lit == "struct" {
		structPos = p.expect(token.IDENT)
		lbrace = p.expect(token.LBRACE)
	}
	var list []*ast.Field
	for p.tok == token.IDENT || p.tok == token.MUL || p.tok == token.LPAREN {
		list = append(list, p.parseFieldDecl())
	}
	rbrace := p.expect(token.RBRACE)

	return &ast.StructType{
		Struct: structPos,
		Fields: &ast.FieldList{
			Lbrace: lbrace,
			List:   list,
			Rbrace: rbrace,
		},
	}
}

func (p *parser) parseFieldDecl() *ast.Field {
	if p.trace {
		defer un(trace(p, "FieldDecl"))
	}

	doc := p.leadComment

	var names []*ast.Ident
	var typ ast.Expr
	if p.tok == token.IDENT {
		name := p.parseIdent(true)
		if p.tok == token.STRING || p.tok == token.SEMICOLON || p.tok == token.RBRACE {
			typ = name
		} else {
			names = []*ast.Ident{name}
			for p.tok == token.COMMA {
				p.next()
				names = append(names, p.parseIdent(true))
			}
			typ = p.parseType()
		}
	} else {
		/* type User { map[string]string } */
		typ = p.parseType()
	}

	var tag *ast.BasicLit
	if p.tok == token.STRING {
		tag = &ast.BasicLit{
			ValuePos: p.pos,
			Kind:     p.tok,
			Value:    p.lit,
		}
		p.next()
	}

	p.expectSemi()
	return &ast.Field{
		Doc:     doc,
		Names:   names,
		Type:    typ,
		Tag:     tag,
		Comment: p.lineComment,
	}
}

func (p *parser) parseMapType() *ast.MapType {
	if p.trace {
		defer un(trace(p, "MapType"))
	}

	pos := p.expect(token.IDENT) // map
	p.expect(token.LBRACK)
	key := p.parseType()
	p.expect(token.RBRACK)
	value := p.parseType()

	return &ast.MapType{
		Map:   pos,
		Key:   key,
		Value: value,
	}
}

func (p *parser) parseArrayType() *ast.ArrayType {
	if p.trace {
		defer un(trace(p, "ArrayType"))
	}

	lbrack := p.expect(token.LBRACK)
	p.expect(token.RBRACK)
	elt := p.parseType()
	return &ast.ArrayType{
		Lbrack: lbrack,
		//Len:    nil,
		Elt: elt,
	}
}

func (p *parser) parsePointerType() *ast.StarExpr {
	if p.trace {
		defer un(trace(p, "PointerType"))
	}

	star := p.expect(token.MUL)
	base := p.parseType()

	return &ast.StarExpr{
		Star: star,
		X:    base,
	}
}

func (p *parser) parseParenExpr() *ast.ParenExpr {
	if p.trace {
		defer un(trace(p, "ParenExpr"))
	}

	lparen := p.expect(token.LPAREN)
	typ := p.parseType()
	rparen := p.expect(token.RPAREN)
	return &ast.ParenExpr{
		Lparen: lparen,
		X:      typ,
		Rparen: rparen,
	}
}

func (p *parser) parseType() ast.Expr {
	if p.trace {
		defer un(trace(p, "Type"))
	}

	typ := p.tryIdentOrType()

	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		p.advance()
		return &ast.BadExpr{From: pos, To: p.pos}
	}
	return typ
}

func (p *parser) tryIdentOrType() ast.Expr {
	switch p.tok {
	case token.IDENT:
		switch p.lit {
		case "struct":
			return p.parseStructType()
		case "map":
			return p.parseMapType()
		default:
			return p.parseIdent(true)
		}
	case token.LBRACE:
		return p.parseStructType()
	case token.LBRACK:
		return p.parseArrayType()
	case token.MUL:
		return p.parsePointerType()
	case token.LPAREN:
		return p.parseParenExpr()
	}
	return nil
}

// ----------------------------------------------------------------------------
// service

func (p *parser) parseService() ast.Decl {
	if p.trace {
		defer un(trace(p, "ServiceDecl"))
	}

	var serviceExt *ast.ServiceExtDecl
	if ast.SERVEREXT.Is(p.lit) {
		serviceExt = p.parseServiceExtDecl()
	}
	if !ast.SERVICE.Is(p.lit) {
		pos := p.pos
		p.errorExpected(pos, "service")
		p.advance()
		return &ast.BadDecl{
			From: pos,
			To:   p.pos,
		}
	}
	serviceApi := p.parseServiceApiDecl()
	p.expectSemi()
	return &ast.ServiceDecl{
		ServiceExt: serviceExt,
		ServiceApi: serviceApi,
	}
}

func (p *parser) parseServiceExtDecl() *ast.ServiceExtDecl {
	if p.trace {
		defer un(trace(p, "ServiceExtDecl"))
	}
	pos := p.expect(token.IDENT)
	lparen := p.expect(token.LPAREN)
	kvs := p.parseElementList()
	rparen := p.expect(token.RPAREN)
	p.expectSemi()

	return &ast.ServiceExtDecl{
		TokPos: pos,
		Lparen: lparen,
		Kvs:    kvs,
		Rparen: rparen,
	}
}

func (p *parser) parseServiceApiDecl() *ast.ServiceApiDecl {
	if p.trace {
		defer un(trace(p, "ServiceApiDecl"))
	}
	pos := p.expect(token.IDENT)
	name := p.parseIdent(false)
	lbrace := p.expect(token.LBRACE)
	serviceRoutes := p.parseServiceRouteList()
	rbrace := p.expect(token.RBRACE)

	return &ast.ServiceApiDecl{
		TokPos:        pos,
		Name:          name,
		Lbrace:        lbrace,
		ServiceRoutes: serviceRoutes,
		Rbrace:        rbrace,
	}
}

func (p *parser) parseServiceRouteList() []*ast.ServiceRouteDecl {
	if p.trace {
		defer un(trace(p, "ServiceRouteList"))
	}

	var routes []*ast.ServiceRouteDecl
	for p.tok != token.RBRACE && p.tok != token.EOF {
		routes = append(routes, p.parseServiceRouteDecl())
	}
	return routes
}

func (p *parser) parseServiceRouteDecl() *ast.ServiceRouteDecl {
	if p.trace {
		defer un(trace(p, "ServiceRouteDecl"))
	}

	pos := p.pos
	var atDoc, atHandler *ast.KeyValueExpr
	for p.tok == token.IDENT {
		if ast.RouteDoc.Is(p.lit) {
			atDoc = p.parseElement(false)
		} else if ast.RouteHandler.Is(p.lit) {
			atHandler = p.parseElement(false)
		} else {
			break
		}
		p.expectSemi()
	}
	route := p.parseRoute()
	return &ast.ServiceRouteDecl{
		TokPos:    pos,
		AtDoc:     atDoc,
		AtHandler: atHandler,
		Route:     route,
	}
}

func (p *parser) parseRoute() *ast.Route {
	if p.trace {
		defer un(trace(p, "Route"))
	}
	method := p.parseIdent(true)
	path := p.parseIdent(false)

	var req, resp *ast.ParenExpr
	if p.tok == token.LPAREN {
		req = p.parseParenExpr()
	}

	var returns token.Pos
	if p.tok == token.IDENT && ast.RouteReturns.Is(p.lit) {
		returns = p.expect(token.IDENT)
		resp = p.parseParenExpr()
	}
	end := p.pos
	p.expectSemi()

	return &ast.Route{
		Method:    method,
		Path:      path,
		Req:       req,
		ReturnPos: returns,
		Resp:      resp,
		EndPos:    end,
	}
}
