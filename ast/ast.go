package ast

import (
	"go/ast"

	"github.com/zeromicro/zero-api/token"
)

type (
	Node = ast.Node

	Expr interface {
		Node
		exprNode()
	}

	Decl interface {
		Node
		declNode()
	}
)

// ----------------------------------------------------------------------------
// comment

type (
	Comment = ast.Comment

	CommentGroup = ast.CommentGroup
)

// ----------------------------------------------------------------------------
// Expressions and types

type (
	BadExpr struct {
		From, To token.Pos
	}

	Ident struct {
		NamePos token.Pos
		Name    string
	}

	// BasicLit node represents a literal of basic type.
	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // token.STRING or token.Ident
		Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A StarExpr node represents an expression of the form "*" Expression.
	// Semantically it could be a unary "*" expression, or a pointer type.
	//
	StarExpr struct {
		Star token.Pos // position of "*"
		X    Expr      // operand
	}

	KeyValueExpr struct {
		Key   *Ident
		Colon token.Pos // position of ":"
		Value *BasicLit // *BasicLit for info or *Ident for server
	}

	// A ParenExpr node represents a parenthesized expression.
	// like (req) or (resp)
	ParenExpr struct {
		Lparen token.Pos // position of "("
		X      Expr      // parenthesized expression
		Rparen token.Pos // position of ")"
	}
)

func (x *BadExpr) Pos() token.Pos { return x.From }
func (x *BadExpr) End() token.Pos { return x.To }
func (x *BadExpr) exprNode()      {}

func (x *Ident) Pos() token.Pos { return x.NamePos }
func (x *Ident) End() token.Pos { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *Ident) exprNode()      {}

func (x *BasicLit) Pos() token.Pos { return x.ValuePos }
func (x *BasicLit) End() token.Pos { return token.Pos(int(x.ValuePos) + len(x.Value)) }
func (x *BasicLit) exprNode()      {}

func (x *StarExpr) Pos() token.Pos { return x.Star }
func (x *StarExpr) End() token.Pos { return x.X.End() }
func (x *StarExpr) exprNode()      {}

func (x *KeyValueExpr) Pos() token.Pos { return x.Key.Pos() }
func (x *KeyValueExpr) End() token.Pos { return x.Value.End() }
func (x *KeyValueExpr) exprNode()      {}

func (x *ParenExpr) Pos() token.Pos { return x.Lparen }
func (x *ParenExpr) End() token.Pos { return x.Rparen }
func (x *ParenExpr) exprNode()      {}

type (
	// A Field represents a Field declaration list in a struct type,
	// a method list in an interface type, or a parameter/result declaration
	// in a signature.
	// Field.Names is nil for unnamed parameters (parameter lists which only contain types)
	// and embedded struct fields. In the latter case, the field name is the type name.
	// Field.Names contains a single name "type" for elements of interface type lists.
	// Types belonging to the same type list share the same "type" identifier which also
	// records the position of that keyword.
	//
	Field struct {
		Doc     *CommentGroup // associated documentation; or nil
		Names   []*Ident      // field/method/(type) parameter names, or type "type"; or nil
		Type    Expr          // field/method/parameter type, type list type; or nil
		Tag     *BasicLit     // field tag; or nil
		Comment *CommentGroup // line comments; or nil
	}

	FieldList struct {
		Lbrace token.Pos
		List   []*Field
		Rbrace token.Pos
	}

	// An ArrayType node represents an array or slice type.
	ArrayType struct {
		Lbrack token.Pos // position of "["
		//Len    Expr      // Ellipsis node for [...]T array types, nil for slice types
		Elt Expr // element type
	}

	StructType struct {
		Struct token.Pos
		Fields *FieldList
		//Incomplete bool
	}

	// A MapType node represents a map type.
	MapType struct {
		Map   token.Pos // position of "map" keyword
		Key   Expr
		Value Expr
	}
)

func (f *FieldList) Pos() token.Pos {
	if f.Lbrace.IsValid() {
		return f.Lbrace
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if len(f.List) > 0 {
		return f.List[0].Pos()
	}
	return token.NoPos
}

func (f *FieldList) End() token.Pos {
	if f.Rbrace.IsValid() {
		return f.Rbrace + 1
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if n := len(f.List); n > 0 {
		return f.List[n-1].End()
	}
	return token.NoPos
}

func (f *Field) Pos() token.Pos {
	if len(f.Names) > 0 {
		return f.Names[0].Pos()
	}
	if f.Type != nil {
		return f.Type.Pos()
	}
	return token.NoPos
}

func (f *Field) End() token.Pos {
	if f.Tag != nil {
		return f.Tag.End()
	}
	if f.Type != nil {
		return f.Type.End()
	}
	if len(f.Names) > 0 {
		return f.Names[len(f.Names)-1].End()
	}
	return token.NoPos
}

// NumFields returns the number of parameters or struct fields represented by a FieldList.
func (f *FieldList) NumFields() int {
	n := 0
	if f != nil {
		for _, g := range f.List {
			m := len(g.Names)
			if m == 0 {
				m = 1
			}
			n += m
		}
	}
	return n
}

func (x *ArrayType) Pos() token.Pos { return x.Lbrack }
func (x *ArrayType) End() token.Pos { return x.Elt.End() }
func (x *ArrayType) exprNode()      {}

func (x *StructType) Pos() token.Pos {
	if x.Struct.IsValid() {
		return x.Struct
	}
	return x.Fields.Pos()
}
func (x *StructType) End() token.Pos { return x.Fields.End() }
func (x *StructType) exprNode()      {}

func (x *MapType) Pos() token.Pos { return x.Map }
func (x *MapType) End() token.Pos { return x.Value.End() }
func (x *MapType) exprNode()      {}

// ----------------------------------------------------------------------------
// spec

type (
	// The Spec type stands for any of *ImportSpec, and *TypeSpec.
	Spec interface {
		Node
		specNode()
	}

	// An ImportSpec node represents a single package import.
	ImportSpec struct {
		Doc *CommentGroup // associated documentation; or nil
		//Name    *Ident        // local package name (including "."); or nil
		Path    *BasicLit     // import path
		Comment *CommentGroup // line comments; or nil
		EndPos  token.Pos     // end of spec (overrides Path.Pos if nonzero)
	}

	// A TypeSpec node represents a type declaration (TypeSpec production).
	TypeSpec struct {
		Doc  *CommentGroup // associated documentation; or nil
		Name *Ident        // type name
		//TypeParams *FieldList    // type parameters; or nil
		//Assign     token.Pos     // position of '=', if any
		Type    *StructType   // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
		Comment *CommentGroup // line comments; or nil
	}
)

func (x *ImportSpec) Pos() token.Pos { return x.Path.Pos() }
func (x *ImportSpec) End() token.Pos { return x.EndPos }
func (x *ImportSpec) specNode()      {}

func (x *TypeSpec) Pos() token.Pos { return x.Name.Pos() }
func (x *TypeSpec) End() token.Pos { return x.Type.Pos() }
func (x *TypeSpec) specNode()      {}

// ----------------------------------------------------------------------------
// decl

type (
	BadDecl struct {
		From, To token.Pos
	}

	SyntaxDecl struct {
		TokPos     token.Pos
		Assign     token.Pos
		SyntaxName *BasicLit
	}

	GenDecl struct {
		Doc    *CommentGroup
		TokPos token.Pos
		Key    Keyword // import, type
		Lparen token.Pos
		Specs  []Spec
		Rparen token.Pos
	}

	InfoDecl struct {
		TokPos   token.Pos
		Lparen   token.Pos
		Elements []*KeyValueExpr
		Rparen   token.Pos
	}
)

func (x *BadDecl) Pos() token.Pos { return x.From }
func (x *BadDecl) End() token.Pos { return x.To }
func (x *BadDecl) declNode()      {}

func (x *GenDecl) Pos() token.Pos { return x.TokPos }
func (x *GenDecl) End() token.Pos {
	if x.Rparen.IsValid() {
		return x.Rparen + 1
	}
	return x.Specs[0].End()
}
func (x *GenDecl) declNode() {}

// ----------------------------------------------------------------------------
// service

type (
	ServiceDecl struct {
		ServiceExt *ServiceExtDecl
		ServiceApi *ServiceApiDecl
	}

	ServiceExtDecl struct {
		TokPos token.Pos // @server pos
		Lparen token.Pos
		Kvs    []*KeyValueExpr
		Rparen token.Pos
	}

	ServiceApiDecl struct {
		TokPos        token.Pos
		Name          *Ident
		Lbrace        token.Pos
		ServiceRoutes []*ServiceRouteDecl
		Rbrace        token.Pos
	}

	ServiceRouteDecl struct {
		TokPos    token.Pos
		AtDoc     *KeyValueExpr
		AtHandler *KeyValueExpr
		Route     *Route
	}

	Route struct {
		Method    *Ident
		Path      *Ident
		Req       *ParenExpr
		ReturnPos token.Pos // returns pos
		Resp      *ParenExpr
		EndPos    token.Pos // because Resp Req is optional, need this for EndPos
	}
)

func (x *ServiceDecl) Pos() token.Pos {
	if x.ServiceExt != nil {
		return x.ServiceExt.Pos()
	}
	return x.ServiceApi.Pos()
}
func (x *ServiceDecl) End() token.Pos {
	return x.ServiceApi.End()
}
func (x *ServiceDecl) declNode() {}

func (x *ServiceExtDecl) Pos() token.Pos   { return x.TokPos }
func (x *ServiceExtDecl) token() token.Pos { return x.Rparen }

func (x *ServiceApiDecl) Pos() token.Pos { return x.TokPos }
func (x *ServiceApiDecl) End() token.Pos { return x.Rbrace }

// ----------------------------------------------------------------------------
// File

type File struct {
	Doc *CommentGroup

	SyntaxDecl  *SyntaxDecl
	ImportDecls []*GenDecl
	InfoDecl    *InfoDecl
	Decls       []Decl // top-level declarations; or nil; types or service
}

// ----------------------------------------------------------------------------
// keyword

type Keyword string

const (
	SYNTAX       Keyword = "syntax"
	IMPORT       Keyword = "import"
	INFO         Keyword = "info"
	TYPE         Keyword = "type"
	SERVICE      Keyword = "service"
	SERVEREXT    Keyword = "@server"
	RouteDoc     Keyword = "@doc"
	RouteHandler Keyword = "@handler"
	RouteReturns Keyword = "returns"
)

func (k Keyword) Is(str string) bool {
	return string(k) == str
}
