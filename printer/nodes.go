package printer

import (
	"github.com/zeromicro/zero-api/ast"
	"github.com/zeromicro/zero-api/token"
)

// setComment sets g as the next comment if g != nil and if node comments
// are enabled - this mode is used when printing source code fragments such
// as exports only. It assumes that there is no pending comment in p.comments
// and at most one pending comment in the p.comment cache.
func (p *printer) setComment(g *ast.CommentGroup) {
	if g == nil || !p.useNodeComments {
		return
	}
	if p.comments == nil {
		// initialize p.comments lazily
		p.comments = make([]*ast.CommentGroup, 1)
	} else if p.cindex < len(p.comments) {
		// for some reason there are pending comments; this
		// should never happen - handle gracefully and flush
		// all comments up to g, ignore anything after that
		p.flush(p.posFor(g.List[0].Pos()), token.ILLEGAL)
		p.comments = p.comments[0:1]
		// in debug mode, report error
		p.internalError("setComment found pending comments")
	}
	p.comments[0] = g
	p.cindex = 0
	// don't overwrite any pending comment in the p.comment cache
	// (there may be a pending comment when a line comment is
	// immediately followed by a lead comment with no other
	// tokens between)
	if p.commentOffset == infinity {
		p.nextComment() // get comment ready for use
	}
}

func (p *printer) file(node *ast.File) {
	p.setComment(node.Doc)

	p.syntax(node.SyntaxDecl)
	p.infoDecl(node.InfoDecl)
	p.importDecls(node.ImportDecls)
	p.declList(node.Decls)

	p.print(newline)
}

func (p *printer) syntax(node *ast.SyntaxDecl) {
	if node == nil {
		return
	}
	p.print(node.Pos(), ast.SYNTAX, blank)
	p.print(node.Assign, token.ASSIGN, blank)
	p.expr(node.SyntaxName)
	p.print(newline)
}

func (p *printer) infoDecl(node *ast.InfoDecl) {
	if node == nil {
		return
	}
	if len(p.output) > 0 {
		p.print(newline)
	}

	p.print(node.Pos(), ast.INFO, blank, token.LPAREN, newline)
	p.print(indent)
	for _, each := range node.Elements {
		p.expr(each)
		p.print(newline)
	}
	p.print(unindent, token.RPAREN, newline)
}

func (p *printer) importDecls(nodes []*ast.GenDecl) {
	for _, each := range nodes {
		if len(p.output) > 0 {
			p.print(newline)
		}
		p.genDecl(each)
	}
}

// genDecl for import or type
func (p *printer) genDecl(node *ast.GenDecl) {
	p.print(node.Pos(), node.Key, blank)
	if len(node.Specs) == 0 {
		p.print(node.Lparen, token.LPAREN)
		p.print(node.Rparen, token.RPAREN)
		return
	}

	if node.Lparen.IsValid() || len(node.Specs) > 1 {
		p.print(node.Lparen, token.LPAREN)
		p.print(indent, formfeed)
		for i, s := range node.Specs {
			if i > 0 {
				p.print(newline)
				if node.Key == ast.TYPE {
					p.print(newline)
				}
			}
			p.spec(s)
		}
		p.print(unindent, formfeed)
		p.print(node.Rparen, token.RPAREN)
	} else if len(node.Specs) > 0 { // one line declaration
		p.spec(node.Specs[0])
	}
}

// spec for importSpec or typeSpec
func (p *printer) spec(spec ast.Spec) {
	switch x := spec.(type) {
	case *ast.ImportSpec:
		p.expr(x.Path)
		p.print(x.End())

	case *ast.TypeSpec:
		p.expr(x.Name)
		p.expr(x.Type)
		p.print(x.End())

	default:
		panic("unreachable")
	}
}

func (p *printer) expr(node ast.Expr) {
	p.print(node.Pos())
	switch x := node.(type) {
	case *ast.BadExpr:
		// todo:
	case *ast.Ident:
		p.print(x)

	case *ast.BasicLit:
		p.print(x)

	case *ast.KeyValueExpr:
		p.print(x.Key)
		if x.Colon.IsValid() {
			p.print(token.COLON)
		}
		p.print(blank, x.Value)

	case *ast.ParenExpr:
		p.print(token.LPAREN, x.X, token.RPAREN)

	case *ast.ArrayType:
		p.print(token.LBRACK)
		p.expr(x.Elt)
		p.print(token.RBRACK)

	case *ast.StructType:
		p.print(ast.TYPE)
		p.fieldList(x.Fields)

	case *ast.MapType:
		p.print("map", token.LBRACK)
		p.expr(x.Key)
		p.print(token.RBRACK)
		p.expr(x.Value)

	default:
		panic("unreachable")
	}
}

func (p *printer) fieldList(fields *ast.FieldList) {
	if fields == nil {
		return
	}
	p.print(fields.Pos(), blank, fields.Lbrace, token.LBRACE, indent)
	// TODO: has comment hasComments ||
	if len(fields.List) > 0 {
		p.print(formfeed)
	}
	sep := vtab
	if len(fields.List) == 1 {
		sep = blank
	}
	for i, f := range fields.List {
		if i > 0 {
			p.print(newline)
		}
		for j, x := range f.Names {
			if j > 0 {
				p.print(x.Pos(), token.COMMA, blank)
			}
			p.expr(x)
		}
		p.print(sep)
		p.print(f.Type)
		if f.Tag != nil {
			p.print(sep)
			p.expr(f.Tag)
		}
	}
	p.print(unindent, formfeed, fields.Rbrace, token.RBRACE)
}

// ----------------------------------------------------------------------------
// decl

func (p *printer) declList(decls []ast.Decl) {
	for _, d := range decls {
		p.print(newline, newline)
		p.decl(d)
	}
}

func (p *printer) decl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.BadDecl:
		p.print(d.Pos(), "BadDecl")

	case *ast.GenDecl: // just type decl
		p.genDecl(d)

	case *ast.ServiceDecl:
		p.serviceDecl(d)

	default:
		panic("unreachable")
	}
}

// ----------------------------------------------------------------------------
// service decl

func (p *printer) serviceDecl(node *ast.ServiceDecl) {
	if node.ServiceExt != nil {
		p.serviceExtDecl(node.ServiceExt)
		p.print(newline)
	}

	p.serviceApiDecl(node.ServiceApi)
}

func (p *printer) serviceExtDecl(node *ast.ServiceExtDecl) {
	p.print(node.Pos(), ast.SERVEREXT, blank, token.LPAREN)
	p.print(indent, formfeed)
	for _, each := range node.Kvs {
		//p.print(each.Pos())
		p.print(newline)
		p.expr(each)
	}
	p.print(unindent, formfeed, token.RPAREN)
}

func (p *printer) serviceApiDecl(node *ast.ServiceApiDecl) {
	p.print(node.Pos(), ast.SERVICE, blank)
	p.expr(node.Name)
	p.print(blank, node.Lbrace, token.LBRACE)
	p.print(indent, formfeed)
	for i, x := range node.ServiceRoutes {
		if i > 0 {
			p.print(newline, newline)
		}
		p.serviceRouteDecl(x)
	}
	p.print(unindent, formfeed)
	p.print(token.RBRACE)
}

func (p *printer) serviceRouteDecl(node *ast.ServiceRouteDecl) {
	p.print(node.TokPos)
	if node.AtDoc != nil {
		p.expr(node.AtDoc)
		p.print(newline)
	}
	if node.AtHandler != nil {
		p.expr(node.AtHandler)
		p.print(newline)
	}
	p.route(node.Route)
}

func (p *printer) route(node *ast.Route) {
	p.print(node.Method, blank, node.Path)
	if node.Req != nil {
		p.print(blank)
		p.expr(node.Req)
	}
	if node.Resp != nil {
		p.print(blank)
		p.print(ast.RouteReturns, blank)
		p.expr(node.Resp)
	}
}
